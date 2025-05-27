// src/components/LoadingSpinner.js
import React from 'react';

const LoadingSpinner = () => {
  return (
    <div>
      {/* 这里是加载动画的具体内容 */}
      <i className="fa-solid fa-spinner animate-spin"></i>
    </div>
  );
};

export default LoadingSpinner;