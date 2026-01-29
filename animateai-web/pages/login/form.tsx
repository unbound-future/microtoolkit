import {
  Form,
  Input,
  Checkbox,
  Link,
  Button,
  Space,
} from '@arco-design/web-react';
import { FormInstance } from '@arco-design/web-react/es/Form';
import { IconLock, IconUser } from '@arco-design/web-react/icon';
import React, { useEffect, useRef, useState } from 'react';
import request from '@/utils/request';
import useStorage from '@/utils/useStorage';
import useLocale from '@/utils/useLocale';
import locale from './locale';
import styles from './style/index.module.less';

export default function LoginForm() {
  const formRef = useRef<FormInstance>();
  const [errorMessage, setErrorMessage] = useState('');
  const [loading, setLoading] = useState(false);
  const [loginParams, setLoginParams, removeLoginParams] =
    useStorage('loginParams');

  const t = useLocale(locale);

  // 使用 false 作为初始值，避免 SSR 和客户端不一致
  const [rememberPassword, setRememberPassword] = useState(false);
  const [mounted, setMounted] = useState(false);

  // 在客户端挂载后设置 rememberPassword 的值
  useEffect(() => {
    setMounted(true);
    if (loginParams) {
      setRememberPassword(true);
    }
  }, [loginParams]);

  function afterLoginSuccess(params) {
    // 记住密码
    if (rememberPassword) {
      setLoginParams(JSON.stringify(params));
    } else {
      removeLoginParams();
    }
    // 记录登录状态
    localStorage.setItem('userStatus', 'login');
    // 跳转首页
    window.location.href = '/';
  }

  function login(params) {
    setErrorMessage('');
    setLoading(true);
    console.log('Login function called with params:', params);
    request
      .post('/api/user/login', params)
      .then((res) => {
        console.log('Login response:', res.data);
        const { status, msg } = res.data;
        if (status === 'ok') {
          afterLoginSuccess(params);
        } else {
          setErrorMessage(msg || t['login.form.login.errMsg']);
        }
      })
      .catch((err) => {
        console.error('Login error:', err);
        if (err.response && err.response.data && err.response.data.msg) {
          setErrorMessage(err.response.data.msg);
        } else if (err.message) {
          setErrorMessage(err.message);
        } else {
          setErrorMessage(t['login.form.login.errMsg']);
        }
      })
      .finally(() => {
        setLoading(false);
      });
  }

  function register(params) {
    setErrorMessage('');
    setLoading(true);
    request
      .post('/api/user/register', params)
      .then((res) => {
        const { status, msg } = res.data;
        if (status === 'ok') {
          // 注册成功后自动登录
          // 先重置 loading 状态，然后延迟调用登录
          setLoading(false);
          // 使用 setTimeout 确保数据库事务已提交，增加延迟时间
          // 使用更长的延迟确保数据库事务完全提交
          setTimeout(() => {
            console.log('Register success, calling login after delay...', params);
            login(params);
          }, 1000);
        } else {
          setErrorMessage(msg || t['login.form.register.errMsg']);
          setLoading(false);
        }
      })
      .catch((err) => {
        console.error('Register error:', err);
        if (err.response && err.response.data && err.response.data.msg) {
          setErrorMessage(err.response.data.msg);
        } else if (err.message) {
          setErrorMessage(err.message);
        } else {
          setErrorMessage(t['login.form.register.errMsg']);
        }
        setLoading(false);
      });
  }

  function onRegisterClick() {
    formRef.current.validate().then((values) => {
      register(values);
      });
  }

  function onSubmitClick() {
    formRef.current.validate().then((values) => {
      login(values);
    });
  }

  // 读取 localStorage，设置初始值（只在客户端挂载后执行）
  useEffect(() => {
    if (!mounted) return;
    const shouldRemember = !!loginParams;
    setRememberPassword(shouldRemember);
    if (formRef.current && shouldRemember) {
      const parseParams = JSON.parse(loginParams);
      formRef.current.setFieldsValue(parseParams);
    }
  }, [loginParams, mounted]);

  return (
    <div className={styles['login-form-wrapper']}>
      <div className={styles['login-form-title']}>{t['login.form.title']}</div>
      <div className={styles['login-form-sub-title']}>
        {t['login.form.title']}
      </div>
      <div className={styles['login-form-error-msg']}>{errorMessage}</div>
      <Form
        className={styles['login-form']}
        layout="vertical"
        ref={formRef}
        initialValues={{ userName: 'admin', password: 'admin' }}
      >
        <Form.Item
          field="userName"
          rules={[{ required: true, message: t['login.form.userName.errMsg'] }]}
        >
          <Input
            prefix={<IconUser />}
            placeholder={t['login.form.userName.placeholder']}
            onPressEnter={onSubmitClick}
          />
        </Form.Item>
        <Form.Item
          field="password"
          rules={[{ required: true, message: t['login.form.password.errMsg'] }]}
        >
          <Input.Password
            prefix={<IconLock />}
            placeholder={t['login.form.password.placeholder']}
            onPressEnter={onSubmitClick}
          />
        </Form.Item>
        <Space size={16} direction="vertical">
          <div className={styles['login-form-password-actions']}>
            <Checkbox checked={rememberPassword} onChange={setRememberPassword}>
              {t['login.form.rememberPassword']}
            </Checkbox>
            <Link>{t['login.form.forgetPassword']}</Link>
          </div>
          <Button type="primary" long onClick={onSubmitClick} loading={loading}>
            {t['login.form.login']}
          </Button>
          <Button
            type="text"
            long
            className={styles['login-form-register-btn']}
            onClick={onRegisterClick}
            loading={loading}
          >
            {t['login.form.register']}
          </Button>
        </Space>
      </Form>
    </div>
  );
}
