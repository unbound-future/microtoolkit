import axios from 'axios';

// 创建 axios 实例
const request = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_BASE_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// 请求拦截器
request.interceptors.request.use(
  (config) => {
    // 打印请求信息
    console.log('🚀 Request:', config.method?.toUpperCase(), config.url, config.data);
    
    // 添加 Basic Auth 认证
    if (typeof window !== 'undefined') {
      const loginParams = localStorage.getItem('loginParams');
      if (loginParams) {
        try {
          const params = JSON.parse(loginParams);
          const userName = params.userName;
          const password = params.password;
          if (userName && password) {
            // 创建 Basic Auth header
            const credentials = btoa(`${userName}:${password}`);
            config.headers.Authorization = `Basic ${credentials}`;
          }
        } catch (e) {
          console.error('Failed to parse loginParams for Basic Auth:', e);
        }
      }
    }
    
    // 如果是FormData，不要设置Content-Type，让axios自动设置（包含boundary）
    if (config.data instanceof FormData) {
      delete config.headers['Content-Type'];
    } else if (!config.headers['Content-Type']) {
      // 只有在不是FormData且没有指定Content-Type时才使用默认的application/json
      config.headers['Content-Type'] = 'application/json';
    }
    
    return config;
  },
  (error) => {
    console.error('❌ Request error:', error);
    return Promise.reject(error);
  }
);

// 响应拦截器
request.interceptors.response.use(
  (response) => {
    console.log('✅ Response:', response.config.method?.toUpperCase(), response.config.url, response.status, response.data);
    return response;
  },
  (error) => {
    console.error('❌ Response error:', error.config?.method?.toUpperCase(), error.config?.url, error.response?.status, error.response?.data);
    // 统一错误处理
    return Promise.reject(error);
  }
);

export default request;

