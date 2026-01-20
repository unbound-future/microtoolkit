import React, { useEffect, useState } from 'react';
import {
  Button,
  Avatar,
  Upload,
  Descriptions,
  Tag,
  Skeleton,
  Link,
} from '@arco-design/web-react';
import { IconCamera, IconPlus } from '@arco-design/web-react/icon';
import useLocale from '@/utils/useLocale';
import locale from './locale';
import styles from './style/header.module.less';

export default function Info({
  userInfo = {},
  loading,
}: {
  userInfo: any;
  loading: boolean;
}) {
  const t = useLocale(locale);

  const [avatar, setAvatar] = useState('');

  function onAvatarChange(_, file) {
    setAvatar(file.originFile ? URL.createObjectURL(file.originFile) : '');
  }

  useEffect(() => {
    if (userInfo && userInfo.avatar) {
    setAvatar(userInfo.avatar);
    }
  }, [userInfo]);

  const loadingImg = (
    <Skeleton
      text={{ rows: 0 }}
      style={{ width: '100px', height: '100px' }}
      animation
    />
  );

  const loadingNode = <Skeleton text={{ rows: 1 }} animation />;
  return (
    <div className={styles['info-wrapper']}>
      <Upload showUploadList={false} onChange={onAvatarChange}>
        {loading ? (
          loadingImg
        ) : (
          <Avatar
            size={100}
            triggerIcon={<IconCamera />}
            className={styles['info-avatar']}
          >
            {avatar ? <img src={avatar} /> : <IconPlus />}
          </Avatar>
        )}
      </Upload>
      <Descriptions
        className={styles['info-content']}
        column={1}
        colon="："
        labelStyle={{ textAlign: 'right', width: '120px' }}
        data={[
          {
            label: t['userSetting.label.name'],
            value: loading ? loadingNode : userInfo.name,
          },
          {
            label: t['userSetting.label.accountId'],
            value: loading ? loadingNode : userInfo.accountId,
          },
          {
            label: t['userSetting.label.phoneNumber'],
            value: loading ? loadingNode : userInfo.phoneNumber,
          },
          {
            label: t['userSetting.label.registrationTime'],
            value: loading ? loadingNode : userInfo.registrationTime,
          },
        ]}
      ></Descriptions>
    </div>
  );
}
