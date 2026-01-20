import React, { useContext, useEffect, useState, useMemo } from 'react';
import { useSelector } from 'react-redux';
import useLocale from '@/utils/useLocale';
import locale from './locale';
import { GlobalContext } from '@/context';
import {
  Input,
  Select,
  Button,
  Form,
  Space,
  Message,
  Skeleton,
} from '@arco-design/web-react';
import request from '@/utils/request';
import axios from 'axios';

// 国家数据接口类型
interface CountryData {
  name: {
    common: string;
    official: string;
    nativeName?: {
      [key: string]: {
        common: string;
        official: string;
      };
    };
  };
  cca2: string; // 两位国家代码
  cca3: string; // 三位国家代码
  ccn3?: string; // 数字代码
}

function InfoForm({ loading }: { loading?: boolean }) {
  const t = useLocale(locale);
  const [form] = Form.useForm();
  const { lang } = useContext(GlobalContext);
  const userInfo = useSelector((state: any) => state.userInfo || {});
  
  // 国家数据状态
  const [countriesData, setCountriesData] = useState<CountryData[]>([]);
  const [countriesLoading, setCountriesLoading] = useState(false);
  
  // 从 restcountries.com API 获取所有国家数据
  useEffect(() => {
    const fetchCountries = async () => {
      // 先检查 localStorage 是否有缓存
      const cachedData = localStorage.getItem('countries_data');
      const cacheTime = localStorage.getItem('countries_data_time');
      const now = Date.now();
      
      // 如果缓存存在且未过期（24小时），使用缓存
      if (cachedData && cacheTime && (now - parseInt(cacheTime)) < 24 * 60 * 60 * 1000) {
        try {
          setCountriesData(JSON.parse(cachedData));
          return;
        } catch (e) {
          console.error('Failed to parse cached countries data:', e);
        }
      }
      
      // 从 API 获取数据
      setCountriesLoading(true);
      try {
        const response = await axios.get<CountryData[]>('https://restcountries.com/v3.1/all', {
          params: {
            fields: 'name,cca2,cca3,ccn3', // 只获取需要的字段，减少数据量
          },
        });
        
        // 按国家名称排序
        const sortedData = response.data.sort((a, b) => {
          const nameA = a.name.common.toLowerCase();
          const nameB = b.name.common.toLowerCase();
          return nameA.localeCompare(nameB);
        });
        
        setCountriesData(sortedData);
        
        // 缓存数据
        localStorage.setItem('countries_data', JSON.stringify(sortedData));
        localStorage.setItem('countries_data_time', now.toString());
      } catch (error) {
        console.error('Failed to fetch countries data:', error);
        Message.error('获取国家列表失败，请稍后重试');
      } finally {
        setCountriesLoading(false);
      }
    };
    
    fetchCountries();
  }, []);
  
  // 处理国家数据，转换为选项格式（只使用英文）
  const countriesOptions = useMemo(() => {
    return countriesData.map((country) => {
      return {
        code: country.cca2,
        code3: country.cca3,
        name: country.name.common, // 使用英文名称
        official: country.name.official,
      };
    });
  }, [countriesData]);

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

  // 当 userInfo 加载完成后，填充表单
  useEffect(() => {
    if (!loading && userInfo && Object.keys(userInfo).length > 0) {
      // 处理国家字段：如果后端返回的是国家代码，转换为英文名称
      let rangeAreaValue = userInfo.rangeArea || userInfo.range_area || '';
      
      // 如果 countriesOptions 已加载，尝试匹配国家
      if (rangeAreaValue && countriesOptions.length > 0) {
        // 查找对应的国家对象（支持代码或名称匹配）
        const country = countriesOptions.find(
          c => c.name === rangeAreaValue || c.code === rangeAreaValue || c.code3 === rangeAreaValue
        );
        if (country) {
          // 使用英文名称作为存储值
          rangeAreaValue = country.name;
        }
      }
      // 如果 countriesOptions 还没加载，但 rangeAreaValue 有值，直接使用（可能是英文名称）
      
      const formValues: any = {
        userName: userInfo.userName || '',
        email: userInfo.email || '',
        nickName: userInfo.name || userInfo.userName || '', // 如果没有 name，使用 userName
        phoneNumber: userInfo.phoneNumber || userInfo.phone_number || '',
        rangeArea: rangeAreaValue || '',
        address: userInfo.address || '', // 确保 address 字段被填充
        profile: userInfo.introduction || '',
      };
      
      // 填充表单，确保所有字段都被设置（包括空字符串，这样能正确显示后端返回的数据）
      form.setFieldsValue(formValues);
    }
  }, [userInfo, loading, form, countriesOptions]);
  
  // 当 countriesOptions 加载完成后，如果 userInfo 中有 rangeArea，重新匹配
  useEffect(() => {
    if (!loading && userInfo && Object.keys(userInfo).length > 0 && countriesOptions.length > 0) {
      const currentRangeArea = form.getFieldValue('rangeArea');
      const userRangeArea = userInfo.rangeArea || userInfo.range_area || '';
      
      // 如果表单中的值和用户信息中的值不一致，或者表单中没有值但用户信息中有值
      if (userRangeArea && (!currentRangeArea || currentRangeArea !== userRangeArea)) {
        // 查找对应的国家对象
        const country = countriesOptions.find(
          c => c.name === userRangeArea || c.code === userRangeArea || c.code3 === userRangeArea
        );
        if (country) {
          // 使用英文名称更新表单
          form.setFieldValue('rangeArea', country.name);
        } else if (userRangeArea) {
          // 如果找不到匹配，直接使用原值（可能是英文名称）
          form.setFieldValue('rangeArea', userRangeArea);
        }
      }
    }
  }, [countriesOptions, userInfo, loading, form]);

  const handleSave = async () => {
    try {
      const values = await form.validate();
      const accountId = userInfo.accountId || userInfo.account_id;
      
      if (!accountId) {
        Message.error('无法获取当前登录用户的账户ID，请重新登录');
        return;
      }

      // 将表单数据转换为后端需要的格式
      const saveData = {
        userName: values.userName,
        email: values.email,
        name: values.nickName,
        phoneNumber: values.phoneNumber,
        rangeArea: values.rangeArea,
        address: values.address,
        introduction: values.profile,
      };

      // 调用后端 API，使用账户ID而不是用户名
      const response = await request.post(`/api/user/saveInfo?accountId=${encodeURIComponent(accountId)}`, saveData);
      
      if (response.data.status === 'ok') {
        Message.success(t['userSetting.saveSuccess']);
        // 触发父组件重新获取用户信息（不刷新整个页面）
        if (typeof window !== 'undefined') {
          window.dispatchEvent(new CustomEvent('refreshUserInfo'));
        }
      } else {
        Message.error(response.data.msg || '保存失败');
      }
    } catch (error: any) {
      console.error('Save user info error:', error);
      if (error.response && error.response.data && error.response.data.msg) {
        Message.error(error.response.data.msg);
      } else {
        Message.error('保存失败，请重试');
      }
    }
  };

  const handleReset = () => {
    form.resetFields();
  };

  const loadingNode = (rows = 1) => {
    return (
      <Skeleton
        text={{
          rows,
          width: new Array(rows).fill('100%'),
        }}
        animation
      />
    );
  };

  return (
    <Form
      style={{ width: '500px', marginTop: '6px' }}
      form={form}
      labelCol={{ span: lang === 'en-US' ? 7 : 6 }}
      wrapperCol={{ span: lang === 'en-US' ? 17 : 18 }}
    >
      <Form.Item
        label={t['userSetting.info.userName']}
        field="userName"
        rules={[
          {
            required: true,
            message: t['userSetting.info.userName.placeholder'],
          },
        ]}
      >
        {loading ? (
          loadingNode()
        ) : (
          <Input placeholder={t['userSetting.info.userName.placeholder']} />
        )}
      </Form.Item>
      <Form.Item
        label={t['userSetting.info.email']}
        field="email"
        rules={[
          {
            type: 'email',
            required: true,
            message: t['userSetting.info.email.placeholder'],
          },
        ]}
      >
        {loading ? (
          loadingNode()
        ) : (
          <Input placeholder={t['userSetting.info.email.placeholder']} />
        )}
      </Form.Item>
      <Form.Item
        label={t['userSetting.info.nickName']}
        field="nickName"
        rules={[
          {
            required: true,
            message: t['userSetting.info.nickName.placeholder'],
          },
        ]}
      >
        {loading ? (
          loadingNode()
        ) : (
          <Input placeholder={t['userSetting.info.nickName.placeholder']} />
        )}
      </Form.Item>
      <Form.Item
        label={t['userSetting.info.phoneNumber']}
        field="phoneNumber"
        rules={[
          {
            validator: (value, callback) => {
              if (value && !/^1[3-9]\d{9}$/.test(value)) {
                callback(t['userSetting.info.phoneNumber.placeholder']);
              } else {
                callback();
              }
            },
          },
        ]}
      >
        {loading ? (
          loadingNode()
        ) : (
          <Input placeholder={t['userSetting.info.phoneNumber.placeholder']} />
        )}
      </Form.Item>
      <Form.Item
        label={t['userSetting.info.area']}
        field="rangeArea"
        rules={[
          { required: true, message: t['userSetting.info.area.placeholder'] },
        ]}
      >
        {loading ? (
          loadingNode()
        ) : (
          <Select
            placeholder={t['userSetting.info.area.placeholder']}
            showSearch
            loading={countriesLoading}
            allowClear
            // 本地搜索：使用 filterOption 进行客户端过滤
            // 参考 Arco Design 官方示例：option.props.value 和 option.props.children
            filterOption={(inputValue, option: any) => {
              if (!inputValue) {
                return true;
              }
              const searchText = inputValue.toLowerCase();
              // 使用 option.props.value 和 option.props.children（官方示例的方式）
              const value = String(option?.props?.value || '').toLowerCase();
              const children = String(option?.props?.children || '').toLowerCase();
              return value.indexOf(searchText) >= 0 || children.indexOf(searchText) >= 0;
            }}
          >
            {countriesOptions.map((country) => (
              <Select.Option key={country.code} value={country.name}>
                {country.name}
              </Select.Option>
            ))}
          </Select>
        )}
      </Form.Item>
      <Form.Item label={t['userSetting.info.address']} field="address">
        {loading ? (
          loadingNode()
        ) : (
          <Input placeholder={t['userSetting.info.address.placeholder']} />
        )}
      </Form.Item>
      <Form.Item label={t['userSetting.info.profile']} field="profile">
        {loading ? (
          loadingNode(3)
        ) : (
          <Input.TextArea
            placeholder={t['userSetting.info.profile.placeholder']}
          />
        )}
      </Form.Item>

      <Form.Item label=" ">
        <Space>
          <Button type="primary" onClick={handleSave}>
            {t['userSetting.save']}
          </Button>
          <Button onClick={handleReset}>{t['userSetting.reset']}</Button>
        </Space>
      </Form.Item>
    </Form>
  );
}

export default InfoForm;
