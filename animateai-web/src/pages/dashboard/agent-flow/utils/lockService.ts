// 锁定服务 - 模拟后端接口，后续替换为真实 API

export interface LockInfo {
  locked: boolean;
  lockedBy?: string;
  lockedAt?: number;
  lockId?: string;
}

// 模拟用户ID（实际应该从用户信息中获取）
export const getCurrentUserId = (): string => {
  if (typeof window !== 'undefined') {
    const userInfo = localStorage.getItem('userInfo');
    if (userInfo) {
      try {
        const parsed = JSON.parse(userInfo);
        return parsed.name || parsed.id || 'unknown';
      } catch {
        return 'unknown';
      }
    }
  }
  return 'unknown';
};

// 模拟锁存储（实际应该调用后端 API）
let currentLock: LockInfo | null = null;

/**
 * 尝试获取编辑锁
 * @param flowId 流程ID（这里使用固定ID，实际应该从路由或参数获取）
 * @returns Promise<LockInfo>
 */
export async function acquireLock(flowId?: string): Promise<LockInfo> {
  const currentUserId = getCurrentUserId();
  console.log('[LockService] acquireLock - 请求获取锁', {
    flowId,
    currentUserId,
    currentLockState: currentLock,
  });

  // 模拟 API 调用延迟
  await new Promise((resolve) => setTimeout(resolve, 300));

  // 模拟后端逻辑：如果已经被锁定，返回锁定信息
  if (currentLock?.locked) {
    const result = {
      locked: false,
      lockedBy: currentLock.lockedBy,
      lockedAt: currentLock.lockedAt,
    };
    console.log('[LockService] acquireLock - 获取锁失败，已被锁定', {
      result,
      currentLock,
    });
    return result;
  }

  // 成功获取锁
  const lockInfo: LockInfo = {
    locked: true,
    lockedBy: currentUserId,
    lockedAt: Date.now(),
    lockId: `lock-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
  };

  currentLock = lockInfo;
  console.log('[LockService] acquireLock - 成功获取锁', {
    lockInfo,
    currentLockState: currentLock,
  });

  // TODO: 实际应该调用后端 API
  // return await axios.post('/api/agent-flow/lock', { flowId });

  return lockInfo;
}

/**
 * 释放编辑锁
 * @param lockId 锁ID
 * @returns Promise<boolean>
 */
export async function releaseLock(lockId?: string): Promise<boolean> {
  console.log('[LockService] releaseLock - 请求释放锁', {
    lockId,
    currentLockState: currentLock,
  });

  // 模拟 API 调用延迟
  await new Promise((resolve) => setTimeout(resolve, 200));

  // 检查是否是当前用户持有的锁
  if (currentLock?.locked && currentLock.lockId === lockId) {
    const previousLock = { ...currentLock };
    currentLock = null;
    console.log('[LockService] releaseLock - 成功释放锁', {
      lockId,
      previousLock,
      currentLockState: currentLock,
    });
    return true;
  }

  console.log('[LockService] releaseLock - 释放锁失败，锁ID不匹配或锁不存在', {
    lockId,
    currentLockState: currentLock,
  });

  // TODO: 实际应该调用后端 API
  // return await axios.post('/api/agent-flow/unlock', { lockId });

  return true;
}

/**
 * 检查锁定状态
 * @param flowId 流程ID
 * @returns Promise<LockInfo>
 */
export async function checkLockStatus(flowId?: string): Promise<LockInfo> {
  console.log('[LockService] checkLockStatus - 检查锁状态', {
    flowId,
    currentLockState: currentLock,
  });

  // 模拟 API 调用延迟
  await new Promise((resolve) => setTimeout(resolve, 200));

  if (currentLock?.locked) {
    const result = {
      locked: true,
      lockedBy: currentLock.lockedBy,
      lockedAt: currentLock.lockedAt,
      lockId: currentLock.lockId,
    };
    console.log('[LockService] checkLockStatus - 锁状态：已锁定', {
      result,
      currentLockState: currentLock,
    });
    return result;
  }

  const result = {
    locked: false,
  };
  console.log('[LockService] checkLockStatus - 锁状态：未锁定', {
    result,
    currentLockState: currentLock,
  });

  // TODO: 实际应该调用后端 API
  // return await axios.get(`/api/agent-flow/lock-status?flowId=${flowId}`);

  return result;
}

/**
 * 清除本地锁（用于页面卸载时清理）
 */
export function clearLocalLock(): void {
  console.log('[LockService] clearLocalLock - 清除本地锁', {
    previousLock: currentLock,
  });
  currentLock = null;
  console.log('[LockService] clearLocalLock - 本地锁已清除', {
    currentLockState: currentLock,
  });
}

