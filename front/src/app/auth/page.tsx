import dynamic from 'next/dynamic';

const Auth = dynamic(() => import('@/views/Auth').then((module) => ({ default: module.Auth })));

const AuthPage = (): JSX.Element => {
  return <Auth />;
};

export default AuthPage;
