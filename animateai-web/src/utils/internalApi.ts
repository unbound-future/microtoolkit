import axios from 'axios';

const internalApi = axios.create({
  baseURL: '',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

export default internalApi;
