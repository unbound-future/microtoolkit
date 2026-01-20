import React from 'react';
import { Input, Button, Space, Popconfirm } from '@arco-design/web-react';
import { IconDelete, IconPlus } from '@arco-design/web-react/icon';
import useLocale from '@/utils/useLocale';
import locale from '@/pages/dashboard/agent-flow/locale';
import styles from './style/components.module.less';

interface UpstreamCallDescriptionEditorProps {
  descriptions: string[];
  onChange: (descriptions: string[]) => void;
}

function UpstreamCallDescriptionEditor({
  descriptions = [],
  onChange,
}: UpstreamCallDescriptionEditorProps) {
  const t = useLocale(locale);

  const handleAdd = () => {
    onChange([...descriptions, '']);
  };

  const handleUpdate = (index: number, value: string) => {
    const updated = [...descriptions];
    updated[index] = value;
    onChange(updated);
  };

  const handleDelete = (index: number) => {
    const updated = descriptions.filter((_, i) => i !== index);
    onChange(updated);
  };

  return (
    <div className={styles.upstreamCallDescriptionEditor}>
      <div className={styles.upstreamCallDescriptionHeader}>
        <span>{t['agentFlow.upstreamCallDescriptions'] || '上游调用描述'}</span>
        <Button
          type="text"
          size="small"
          icon={<IconPlus />}
          onClick={handleAdd}
        >
          {t['agentFlow.addDescription'] || '添加描述'}
        </Button>
      </div>
      <div className={styles.descriptionList}>
        {descriptions.map((description, index) => (
          <div key={index} className={styles.descriptionItem}>
            <Space direction="vertical" size="small" style={{ width: '100%' }}>
              <Input.TextArea
                value={description}
                onChange={(value) => handleUpdate(index, value)}
                placeholder={t['agentFlow.upstreamCallDescriptionPlaceholder'] || '输入上游调用描述'}
                style={{ width: '100%' }}
                rows={4}
                autoSize={{ minRows: 4, maxRows: 10 }}
              />
              <div style={{ display: 'flex', justifyContent: 'flex-end' }}>
                <Popconfirm
                  title={t['agentFlow.deleteDescriptionConfirm'] || '确定删除此描述吗？'}
                  onOk={() => handleDelete(index)}
                >
                  <Button
                    type="text"
                    status="danger"
                    icon={<IconDelete />}
                    size="small"
                  >
                    {t['agentFlow.deleteDescription'] || '删除'}
                  </Button>
                </Popconfirm>
              </div>
            </Space>
          </div>
        ))}
        {descriptions.length === 0 && (
          <div className={styles.emptyHint}>
            {t['agentFlow.noDescriptions'] || '暂无描述，点击上方按钮添加'}
          </div>
        )}
      </div>
    </div>
  );
}

export default UpstreamCallDescriptionEditor;

