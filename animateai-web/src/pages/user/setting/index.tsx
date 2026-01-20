import React, { useState, useEffect, useRef } from 'react';
import { useSelector, useDispatch } from 'react-redux';
import { Card, Tabs } from '@arco-design/web-react';
import useLocale from '@/utils/useLocale';
import locale from './locale';
import InfoHeader from './header';
import InfoForm from './info';
import './mock';
import request from '@/utils/request';
import { generatePermission } from '@/routes';

function UserInfo() {
  const t = useLocale(locale);
  const dispatch = useDispatch();
  const userInfo = useSelector((state: any) => state.userInfo);
  const loading = useSelector((state: any) => state.userLoading);
  const [activeTab, setActiveTab] = useState('basic');

  // 从 localStorage 获取当前登录的用户名
  const getCurrentUserName = () => {
    if (typeof window === 'undefined') return '';
    const loginParams = localStorage.getItem('loginParams');
    if (loginParams) {
      try {
        const params = JSON.parse(loginParams);
        return params.userName || '';
      } catch (e) {
        console.error('Failed to parse loginParams:', e);
      }
    }
    return '';
  };

  // 获取用户信息
  const fetchUserInfo = () => {
    // 优先从 Redux store 获取账户ID
    const currentAccountId = userInfo?.accountId || userInfo?.account_id;
    
    // 如果 Redux 中没有账户ID，尝试从 localStorage 获取（向后兼容）
    let accountId = currentAccountId;
    if (!accountId) {
      const loginParams = typeof window !== 'undefined' ? localStorage.getItem('loginParams') : null;
      if (loginParams) {
        try {
          const params = JSON.parse(loginParams);
          accountId = params.accountId || params.account_id;
          // 向后兼容：如果没有账户ID，使用用户名
          if (!accountId) {
            const userName = params.userName;
            if (!userName) {
              console.warn('No accountId or userName found, cannot fetch user info');
              return Promise.resolve();
            }
            dispatch({
              type: 'update-userInfo',
              payload: { userLoading: true },
            });
            return request
              .get(`/api/user/userInfo?userName=${encodeURIComponent(userName)}`)
              .then((res) => {
                const userInfo = {
                  ...res.data,
                  permissions: res.data.permissions || generatePermission('user'),
                };
                dispatch({
                  type: 'update-userInfo',
                  payload: { userInfo, userLoading: false },
                });
                
                // 将账户ID保存到 localStorage，方便后续使用
                if (userInfo.accountId && typeof window !== 'undefined') {
                  try {
                    const params = JSON.parse(loginParams || '{}');
                    params.accountId = userInfo.accountId;
                    localStorage.setItem('loginParams', JSON.stringify(params));
                  } catch (e) {
                    console.error('Failed to update loginParams with accountId:', e);
                  }
                }
              })
              .catch((err) => {
                console.error('Failed to fetch user info:', err);
                dispatch({
                  type: 'update-userInfo',
                  payload: { userInfo: { permissions: generatePermission('user') }, userLoading: false },
                });
              });
          }
        } catch (e) {
          console.error('Failed to parse loginParams:', e);
        }
      }
    }

    if (!accountId) {
      console.warn('No accountId found, cannot fetch user info');
      return Promise.resolve();
    }

    dispatch({
      type: 'update-userInfo',
      payload: { userLoading: true },
    });

    return request
      .get(`/api/user/userInfo?accountId=${encodeURIComponent(accountId)}`)
      .then((res) => {
        // 确保返回的用户信息包含 permissions 字段，默认设置为普通用户权限
        const userInfo = {
          ...res.data,
          permissions: res.data.permissions || generatePermission('user'),
        };
        dispatch({
          type: 'update-userInfo',
          payload: { userInfo, userLoading: false },
        });
        
        // 将账户ID保存到 localStorage，方便后续使用
        if (userInfo.accountId && typeof window !== 'undefined') {
          const loginParams = localStorage.getItem('loginParams');
          if (loginParams) {
            try {
              const params = JSON.parse(loginParams);
              params.accountId = userInfo.accountId;
              localStorage.setItem('loginParams', JSON.stringify(params));
            } catch (e) {
              console.error('Failed to update loginParams with accountId:', e);
            }
          }
        }
      })
      .catch((err) => {
        console.error('Failed to fetch user info:', err);
        // 错误时也设置普通用户权限
        dispatch({
          type: 'update-userInfo',
          payload: { userInfo: { permissions: generatePermission('user') }, userLoading: false },
        });
      });
  };

  // 使用 ref 防止重复请求
  const fetchingRef = useRef(false);

  // 页面加载时主动获取用户信息
  useEffect(() => {
    // 如果已经在加载中，或者已经有用户信息，就不重复请求
    if (loading || fetchingRef.current) {
      return;
    }
    
    // 如果已经有用户信息，就不重复请求（_app.tsx 已经获取过了）
    if (userInfo && Object.keys(userInfo).length > 0) {
      return;
    }

    fetchingRef.current = true;
    fetchUserInfo().finally(() => {
      fetchingRef.current = false;
    });
  }, []); // 只在组件挂载时执行一次

  // 监听刷新用户信息事件（保存成功后触发）
  useEffect(() => {
    const handleRefreshUserInfo = () => {
      if (!fetchingRef.current) {
        fetchingRef.current = true;
        fetchUserInfo().finally(() => {
          fetchingRef.current = false;
        });
      }
    };

    if (typeof window !== 'undefined') {
      window.addEventListener('refreshUserInfo', handleRefreshUserInfo);
      return () => {
        window.removeEventListener('refreshUserInfo', handleRefreshUserInfo);
      };
    }
  }, []);

  return (
    <div>
      <Card style={{ padding: '14px 20px' }}>
        <InfoHeader userInfo={userInfo} loading={loading} />
      </Card>
      <Card style={{ marginTop: '16px' }}>
        <Tabs activeTab={activeTab} onChange={setActiveTab} type="rounded">
          <Tabs.TabPane key="basic" title={t['userSetting.title.basicInfo']}>
            <InfoForm loading={loading} />
          </Tabs.TabPane>
        </Tabs>
      </Card>
    </div>
  );
}

export default UserInfo;
