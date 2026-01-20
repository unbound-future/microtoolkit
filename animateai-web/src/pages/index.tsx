import React, { useEffect } from 'react';
import type { NextPage } from 'next';
import { useRouter } from 'next/router';

const Home: NextPage = () => {
  const router = useRouter();

  useEffect(() => {
    router.replace('/dashboard/asset-management');
  }, [router]);

  return null;
};

export default Home;
