import type { Asset, AssetType } from '../types';

// 文件类型检测
const imageExtensions = /\.(jpg|jpeg|png|gif|bmp|webp|svg)$/i;
const audioExtensions = /\.(mp3|wav|ogg|aac|flac|m4a)$/i;
const videoExtensions = /\.(mp4|webm|ogg|mov|avi|flv|mkv|wmv)$/i;

/**
 * 从 URL 检测资产类型
 */
export function detectAssetTypeFromUrl(url: string): AssetType | null {
  const lowerUrl = url.toLowerCase();
  if (imageExtensions.test(lowerUrl)) {
    return 'image';
  }
  if (audioExtensions.test(lowerUrl)) {
    return 'audio';
  }
  if (videoExtensions.test(lowerUrl)) {
    return 'video';
  }
  return null;
}

/**
 * 从文件检测资产类型
 */
export function detectAssetTypeFromFile(file: File): AssetType | null {
  const fileName = file.name.toLowerCase();
  if (imageExtensions.test(fileName)) {
    return 'image';
  }
  if (audioExtensions.test(fileName)) {
    return 'audio';
  }
  if (videoExtensions.test(fileName)) {
    return 'video';
  }
  // 也可以从 MIME 类型判断
  if (file.type.startsWith('image/')) {
    return 'image';
  }
  if (file.type.startsWith('audio/')) {
    return 'audio';
  }
  if (file.type.startsWith('video/')) {
    return 'video';
  }
  return null;
}

/**
 * 验证 URL 格式
 */
export function isValidUrl(urlString: string): boolean {
  try {
    const url = new URL(urlString);
    return url.protocol === 'http:' || url.protocol === 'https:';
  } catch {
    return false;
  }
}

/**
 * 格式化文件大小
 */
export function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i];
}

/**
 * 生成资产 ID
 */
export function generateAssetId(): string {
  return `asset-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
}

/**
 * 从文件创建对象 URL（用于预览）
 */
export function createFileUrl(file: File): string {
  return URL.createObjectURL(file);
}

/**
 * 释放对象 URL
 */
export function revokeFileUrl(url: string): void {
  URL.revokeObjectURL(url);
}

