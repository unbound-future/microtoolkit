// 全局节点ID管理器
class NodeIdManager {
  private static instance: NodeIdManager;
  private currentId: number = 0;
  private usedIds: Set<string> = new Set();

  private constructor() {
    // 初始化时加载已使用的ID
    this.loadUsedIds();
  }

  public static getInstance(): NodeIdManager {
    if (!NodeIdManager.instance) {
      NodeIdManager.instance = new NodeIdManager();
    }
    return NodeIdManager.instance;
  }

  // 从现有节点中加载已使用的ID
  public loadFromNodes(nodeIds: string[]) {
    nodeIds.forEach((id) => {
      this.usedIds.add(id);
      // 解析ID中的数字部分
      const match = id.match(/node-(\d+)/);
      if (match) {
        const num = parseInt(match[1], 10);
        if (num >= this.currentId) {
          this.currentId = num + 1;
        }
      }
    });
  }

  // 加载已使用的ID（从存储或现有数据中）
  private loadUsedIds() {
    // 可以在这里添加从localStorage或其他存储加载的逻辑
  }

  // 生成新的唯一节点ID
  public generateId(): string {
    let newId: string;
    do {
      newId = `node-${this.currentId}`;
      this.currentId++;
    } while (this.usedIds.has(newId));

    this.usedIds.add(newId);
    return newId;
  }

  // 注册已使用的ID
  public registerId(id: string) {
    this.usedIds.add(id);
    const match = id.match(/node-(\d+)/);
    if (match) {
      const num = parseInt(match[1], 10);
      if (num >= this.currentId) {
        this.currentId = num + 1;
      }
    }
  }

  // 释放ID（节点删除时）
  public releaseId(id: string) {
    this.usedIds.delete(id);
  }

  // 重置（用于测试或重置场景）
  public reset() {
    this.currentId = 0;
    this.usedIds.clear();
  }
}

export default NodeIdManager.getInstance();

