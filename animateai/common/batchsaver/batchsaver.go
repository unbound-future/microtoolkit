package batchsaver

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// ==================== 字段访问器定义 ====================

// fieldAccessor 描述如何从结构体实例中提取某个字段的值
type fieldAccessor struct {
	index []int // reflect.Value.FieldByIndex 所需的索引路径
	isPtr bool  // 该字段本身是否为指针类型
}

// ==================== 通用批量存储器 ====================

// GenericBatchSaver 通用批量存储器
type GenericBatchSaver[T any] struct {
	db            *gorm.DB
	tableName     string
	fields        []string // 数据库字段列表
	uniqueKeys    []string // 唯一约束字段
	batchSize     int
	flushInterval time.Duration

	buffer    []T
	mu        sync.Mutex
	stopChan  chan struct{}
	wg        sync.WaitGroup
	lastFlush time.Time

	// Schema 预计算缓存
	fieldAccessors   map[string]*fieldAccessor // 数据库列名 -> 字段访问器
	orderedAccessors []*fieldAccessor          // 按照 fields 顺序排列的访问器
}

// ==================== 构造函数 ====================

// NewGenericBatchSaver 创建批量处理器
func NewGenericBatchSaver[T any](config Config) (*GenericBatchSaver[T], error) {
	// 类型检查
	db, ok := config.DB.(*gorm.DB)
	if !ok {
		return nil, fmt.Errorf("DB must be *gorm.DB")
	}

	// 推断模型信息
	var zero T
	tableName, fields, _, err := inferModelInfo(db, zero)
	if err != nil {
		return nil, fmt.Errorf("infer model info: %w", err)
	}

	// 使用配置的表名（如果提供）
	if config.TableName != "" {
		tableName = config.TableName
	}

	// 设置默认值
	if config.BatchSize <= 0 {
		config.BatchSize = 1000
	}
	if config.FlushInterval <= 0 {
		config.FlushInterval = 5 * time.Second
	}

	saver := &GenericBatchSaver[T]{
		db:             db,
		tableName:      tableName,
		fields:         fields,
		uniqueKeys:     config.UniqueKeys,
		batchSize:      config.BatchSize,
		buffer:         make([]T, 0, config.BatchSize),
		stopChan:       make(chan struct{}),
		flushInterval:  config.FlushInterval,
		lastFlush:      time.Now(),
		fieldAccessors: make(map[string]*fieldAccessor),
	}

	// 预计算字段访问路径
	if err := saver.buildFieldAccessors(); err != nil {
		return nil, fmt.Errorf("构建字段访问器失败: %w", err)
	}

	// 启动后台刷新协程
	saver.wg.Add(1)
	go saver.autoFlush()

	return saver, nil
}

// ==================== Schema 预计算方法 ====================

// buildFieldAccessors 预计算所有数据库字段的访问路径
func (s *GenericBatchSaver[T]) buildFieldAccessors() error {
	var model T

	// 使用 GORM schema 解析
	schemaObj, err := schema.Parse(&model, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		// 备选方案：回退到反射解析
		return s.buildFieldAccessorsByReflect()
	}

	// 1. 构建 数据库列名 -> fieldAccessor 的映射
	for _, field := range schemaObj.Fields {
		if field.DBName == "" || field.DBName == "deleted_at" {
			continue
		}
		s.fieldAccessors[field.DBName] = &fieldAccessor{
			index: field.StructField.Index,
			isPtr: field.FieldType.Kind() == reflect.Ptr,
		}
	}

	// 2. 根据 s.fields 顺序，构建 orderedAccessors
	s.orderedAccessors = make([]*fieldAccessor, len(s.fields))
	for i, dbName := range s.fields {
		if accessor, ok := s.fieldAccessors[dbName]; ok {
			s.orderedAccessors[i] = accessor
		} else {
			// 数据库列在模型中未找到，填充nil
			s.orderedAccessors[i] = nil
		}
	}

	return nil
}

// buildFieldAccessorsByReflect 备选：通过反射构建访问器
func (s *GenericBatchSaver[T]) buildFieldAccessorsByReflect() error {
	var model T
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// 递归遍历所有字段（包括嵌套的）
	s.fieldAccessors = make(map[string]*fieldAccessor)
	s.traverseStruct(t, nil)

	// 构建 orderedAccessors
	s.orderedAccessors = make([]*fieldAccessor, len(s.fields))
	for i, dbName := range s.fields {
		if accessor, ok := s.fieldAccessors[dbName]; ok {
			s.orderedAccessors[i] = accessor
		} else {
			s.orderedAccessors[i] = nil
		}
	}

	return nil
}

// traverseStruct 递归遍历结构体，收集字段访问路径
func (s *GenericBatchSaver[T]) traverseStruct(t reflect.Type, parentIndex []int) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		// 组合当前字段的完整索引
		index := make([]int, len(parentIndex)+1)
		copy(index, parentIndex)
		index[len(parentIndex)] = i

		// 处理嵌套结构（匿名字段）
		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			s.traverseStruct(field.Type, index)
			continue
		}

		// 获取数据库列名
		dbName := s.getDBNameFromField(field)
		if dbName == "" {
			continue
		}

		// 存储访问器
		s.fieldAccessors[dbName] = &fieldAccessor{
			index: index,
			isPtr: field.Type.Kind() == reflect.Ptr,
		}
	}
}

// getDBNameFromField 从字段标签解析列名
func (s *GenericBatchSaver[T]) getDBNameFromField(field reflect.StructField) string {
	// 从 gorm 标签中解析 "column:xxx"
	gormTag := field.Tag.Get("gorm")
	if gormTag != "" {
		for _, part := range strings.Split(gormTag, ";") {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, "column:") {
				return strings.TrimPrefix(part, "column:")
			}
		}
		if gormTag == "-" {
			return ""
		}
	}
	// 没有明确 column 标签，则使用字段名的下划线形式
	return toSnakeCase(field.Name)
}

// ==================== 核心操作方法 ====================

// Save 保存一条记录
func (s *GenericBatchSaver[T]) Save(item T) error {
	s.mu.Lock()
	s.buffer = append(s.buffer, item)
	shouldFlush := len(s.buffer) >= s.batchSize
	s.mu.Unlock()

	if shouldFlush {
		return s.Flush()
	}
	return nil
}

// Flush 刷新缓冲区到数据库
func (s *GenericBatchSaver[T]) Flush() error {
	s.mu.Lock()
	if len(s.buffer) == 0 {
		s.mu.Unlock()
		return nil
	}

	// 复制并清空缓冲区
	batch := make([]T, len(s.buffer))
	copy(batch, s.buffer)
	s.buffer = s.buffer[:0]
	s.lastFlush = time.Now()
	s.mu.Unlock()

	return s.batchUpsert(batch)
}

// batchUpsert 执行批量插入/更新
func (s *GenericBatchSaver[T]) batchUpsert(items []T) error {
	if len(items) == 0 {
		return nil
	}

	// 构建VALUES部分
	placeholders := make([]string, 0, len(items))
	args := make([]interface{}, 0, len(items)*len(s.fields))

	for _, item := range items {
		// 构建占位符 (?, ?, ?)
		ph := make([]string, len(s.fields))
		for i := range ph {
			ph[i] = "?"
		}
		placeholders = append(placeholders, "("+strings.Join(ph, ",")+")")

		// 提取字段值
		values, err := s.extractValues(item)
		if err != nil {
			return err
		}
		args = append(args, values...)
	}

	// 构建完整SQL
	sql := s.buildUpsertSQL(placeholders)

	// 执行
	return s.db.Exec(sql, args...).Error
}

// extractValues 提取结构体字段值
func (s *GenericBatchSaver[T]) extractValues(item T) ([]interface{}, error) {
	v := reflect.ValueOf(item)
	// 确保我们操作的是结构体值（非指针）
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// 预分配结果切片
	values := make([]interface{}, len(s.orderedAccessors))

	for i, accessor := range s.orderedAccessors {
		if accessor == nil {
			values[i] = nil
			continue
		}

		// 根据预计算的索引路径直接获取字段
		fieldVal := v.FieldByIndex(accessor.index)

		// 安全处理指针字段
		if accessor.isPtr && !fieldVal.IsNil() {
			fieldVal = fieldVal.Elem()
		}

		values[i] = fieldVal.Interface()
	}
	return values, nil
}

// buildUpsertSQL 构建Upsert SQL语句
func (s *GenericBatchSaver[T]) buildUpsertSQL(placeholders []string) string {
	// 构建INSERT部分
	quotedFields := make([]string, len(s.fields))
	for i, f := range s.fields {
		quotedFields[i] = "`" + f + "`"
	}

	sql := fmt.Sprintf(
		"INSERT INTO `%s` (%s) VALUES %s",
		s.tableName,
		strings.Join(quotedFields, ","),
		strings.Join(placeholders, ","),
	)

	// 如果有唯一键，添加ON DUPLICATE KEY UPDATE
	if len(s.uniqueKeys) > 0 && s.supportsUpsert() {
		// 构建UPDATE部分（更新所有非唯一键字段）
		updates := make([]string, 0, len(s.fields)-len(s.uniqueKeys))
		for _, field := range s.fields {
			if !contains(s.uniqueKeys, field) {
				updates = append(updates, fmt.Sprintf("`%s`=VALUES(`%s`)", field, field))
			}
		}

		if len(updates) > 0 {
			sql += " ON DUPLICATE KEY UPDATE " + strings.Join(updates, ",")
		}
	}

	return sql
}

// supportsUpsert 检查是否支持Upsert（MySQL支持）
func (s *GenericBatchSaver[T]) supportsUpsert() bool {
	// 简化处理：假设是MySQL
	return true
}

// ==================== 辅助方法 ====================

// autoFlush 自动刷新协程
func (s *GenericBatchSaver[T]) autoFlush() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.mu.Lock()
			needsFlush := len(s.buffer) > 0 && time.Since(s.lastFlush) > s.flushInterval
			s.mu.Unlock()

			if needsFlush {
				s.Flush()
			}

		case <-s.stopChan:
			return
		}
	}
}

// Close 关闭处理器
func (s *GenericBatchSaver[T]) Close() error {
	close(s.stopChan)
	s.wg.Wait()
	return s.Flush()
}

// Stats 获取统计信息
func (s *GenericBatchSaver[T]) Stats() (bufferSize int, bufferCap int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.buffer), cap(s.buffer)
}

// ==================== 辅助函数 ====================

// inferModelInfo 推断模型信息（带去重逻辑）
func inferModelInfo(db *gorm.DB, model interface{}) (string, []string, string, error) {
	schemaObj, err := schema.Parse(&model, &sync.Map{}, db.NamingStrategy)
	if err != nil {
		// 回退到反射
		return inferByReflection(model)
	}

	tableName := schemaObj.Table
	fields := make([]string, 0, len(schemaObj.Fields))
	seenFields := make(map[string]bool) // 去重map
	var autoIncrement string

	for _, field := range schemaObj.Fields {
		if field.DBName == "" || field.IgnoreMigration {
			continue
		}

		// 跳过软删除字段
		if field.DBName == "deleted_at" {
			continue
		}

		// 去重检查
		if seenFields[field.DBName] {
			continue
		}
		seenFields[field.DBName] = true

		if field.AutoIncrement {
			autoIncrement = field.DBName
		}

		fields = append(fields, field.DBName)
	}

	return tableName, fields, autoIncrement, nil
}

// inferByReflection 通过反射推断
func inferByReflection(model interface{}) (string, []string, string, error) {
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	tableName := strings.ToLower(t.Name()) + "s"
	fields := make([]string, 0, t.NumField())
	seenFields := make(map[string]bool)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// 跳过非导出字段
		if field.PkgPath != "" {
			continue
		}

		// 获取数据库列名
		dbName := getDBNameFromField(field)
		if dbName == "" {
			continue
		}

		// 去重检查
		if seenFields[dbName] {
			continue
		}
		seenFields[dbName] = true

		fields = append(fields, dbName)
	}

	return tableName, fields, "", nil
}

// getDBNameFromField 从字段获取数据库列名（独立函数）
func getDBNameFromField(field reflect.StructField) string {
	gormTag := field.Tag.Get("gorm")
	if gormTag != "" {
		parts := strings.Split(gormTag, ";")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, "column:") {
				return strings.TrimPrefix(part, "column:")
			}
		}

		if gormTag == "-" {
			return ""
		}
	}

	return toSnakeCase(field.Name)
}

// toSnakeCase 驼峰转下划线
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteByte('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// contains 检查字符串是否在切片中
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
