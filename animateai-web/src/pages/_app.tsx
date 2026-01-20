import React, { useEffect, useMemo, useRef } from 'react';
import { useRouter } from 'next/router';
import { isSSR } from '@/utils/is';
import cookies from 'next-cookies';
import Head from 'next/head';
import type { AppProps } from 'next/app';
import { createStore } from 'redux';
import { Provider } from 'react-redux';
import '../style/global.less';
import { ConfigProvider } from '@arco-design/web-react';
import zhCN from '@arco-design/web-react/es/locale/zh-CN';
import enUS from '@arco-design/web-react/es/locale/en-US';
import axios from 'axios';
import request from '@/utils/request';
import NProgress from 'nprogress';
import rootReducer from '../store';
import { GlobalContext } from '../context';
import checkLogin from '@/utils/checkLogin';
import changeTheme from '@/utils/changeTheme';
import useStorage from '@/utils/useStorage';
import Layout from './layout';
import '../mock';
import { generatePermission } from '../routes';

const store = createStore(rootReducer);

interface RenderConfig {
  arcoLang?: string;
  arcoTheme?: string;
}

export default function MyApp({
  pageProps,
  Component,
  renderConfig,
}: AppProps & { renderConfig: RenderConfig }) {
  const { arcoLang, arcoTheme } = renderConfig;
  const [lang, setLang] = useStorage('arco-lang', arcoLang || 'en-US');
  const [theme, setTheme] = useStorage('arco-theme', arcoTheme || 'light');
  
  // useRouter 必须在组件顶层调用，但在服务端会返回 null
  const router = useRouter();

  const locale = useMemo(() => {
    switch (lang) {
      case 'zh-CN':
        return zhCN;
      case 'en-US':
        return enUS;
      default:
        return enUS;
    }
  }, [lang]);

  function fetchUserInfo() {
    // 优先从 Redux store 获取账户ID，如果没有则从 localStorage 获取
    const currentUserInfo = store.getState().userInfo;
    let accountId = currentUserInfo?.accountId || currentUserInfo?.account_id;
    
    // 如果 Redux 中没有账户ID，尝试从 localStorage 获取（向后兼容）
    if (!accountId) {
      const loginParams = typeof window !== 'undefined' ? localStorage.getItem('loginParams') : null;
      if (loginParams) {
        try {
          const params = JSON.parse(loginParams);
          // 如果登录参数中有账户ID，使用账户ID；否则使用用户名（向后兼容）
          accountId = params.accountId || params.account_id;
          if (!accountId && params.userName) {
            // 向后兼容：如果没有账户ID，使用用户名
            const userName = params.userName;
            store.dispatch({
              type: 'update-userInfo',
              payload: { userLoading: true },
            });
            request.get(`/api/user/userInfo?userName=${encodeURIComponent(userName)}`).then((res) => {
              const userInfo = {
                ...res.data,
                permissions: res.data.permissions || generatePermission('user'),
              };
              store.dispatch({
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
            }).catch((err) => {
              console.error('Failed to fetch user info:', err);
              store.dispatch({
                type: 'update-userInfo',
                payload: { userInfo: { permissions: generatePermission('user') }, userLoading: false },
              });
            });
            return;
          }
        } catch (e) {
          console.error('Failed to parse loginParams:', e);
        }
      }
    }

    if (!accountId) {
      console.warn('No accountId found, cannot fetch user info');
      return;
    }

    store.dispatch({
      type: 'update-userInfo',
      payload: { userLoading: true },
    });
    request.get(`/api/user/userInfo?accountId=${encodeURIComponent(accountId)}`).then((res) => {
      // 确保返回的用户信息包含 permissions 字段，默认设置为普通用户权限
      const userInfo = {
        ...res.data,
        permissions: res.data.permissions || generatePermission('user'),
      };
      store.dispatch({
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
    }).catch((err) => {
      console.error('Failed to fetch user info:', err);
      // 错误时也设置普通用户权限
      store.dispatch({
        type: 'update-userInfo',
        payload: { userInfo: { permissions: generatePermission('user') }, userLoading: false },
      });
    });
  }

  // 使用 ref 防止重复请求
  const fetchingRef = useRef(false);

  useEffect(() => {
    if (checkLogin()) {
      // 防止重复请求
      if (!fetchingRef.current) {
        fetchingRef.current = true;
      fetchUserInfo();
        // 设置一个延迟，允许后续请求
        setTimeout(() => {
          fetchingRef.current = false;
        }, 1000);
      }
    } else if (window.location.pathname.replace(/\//g, '') !== 'login') {
      window.location.pathname = '/login';
    }
  }, []);

  useEffect(() => {
    // 只在客户端执行 router 相关逻辑
    if (isSSR || !router || !router.events) return;

    const handleStart = () => {
      NProgress.set(0.4);
      NProgress.start();
    };

    const handleStop = () => {
      NProgress.done();
    };

    router.events.on('routeChangeStart', handleStart);
    router.events.on('routeChangeComplete', handleStop);
    router.events.on('routeChangeError', handleStop);

    return () => {
      router.events.off('routeChangeStart', handleStart);
      router.events.off('routeChangeComplete', handleStop);
      router.events.off('routeChangeError', handleStop);
    };
  }, [router]);

  useEffect(() => {
    document.cookie = `arco-lang=${lang}; path=/`;
    document.cookie = `arco-theme=${theme}; path=/`;
    changeTheme(theme);
  }, [lang, theme]);

  const contextValue = {
    lang,
    setLang,
    theme,
    setTheme,
  };

  return (
    <>
      <Head>
        <link
          rel="shortcut icon"
          type="image/x-icon"
          href="https://unpkg.byted-static.com/latest/byted/arco-config/assets/favicon.ico"
        />
      </Head>
      <ConfigProvider
        locale={locale}
        componentConfig={{
          Card: {
            bordered: false,
          },
          List: {
            bordered: false,
          },
          Table: {
            border: false,
          },
        }}
      >
        <Provider store={store}>
          <GlobalContext.Provider value={contextValue}>
            {Component.displayName === 'LoginPage' ? (
              <Component {...pageProps} suppressHydrationWarning />
            ) : (
              <Layout>
                <Component {...pageProps} suppressHydrationWarning />
              </Layout>
            )}
          </GlobalContext.Provider>
        </Provider>
      </ConfigProvider>
    </>
  );
}

// fix: next build ssr can't attach the localstorage
MyApp.getInitialProps = async (appContext) => {
  const { ctx } = appContext;
  const serverCookies = cookies(ctx);
  return {
    renderConfig: {
      arcoLang: serverCookies['arco-lang'],
      arcoTheme: serverCookies['arco-theme'],
    },
  };
};
