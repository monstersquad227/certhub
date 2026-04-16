/**
 * Certificate-related utility functions
 */

export type CertificateStatus = 'valid' | 'expiring' | 'expired' | 'pending' | 'failed';
export type CertificateType = 'wildcard' | 'single';

/**
 * Get status tag color based on certificate status
 */
export function getStatusColor(status: CertificateStatus): string {
  const colorMap: Record<CertificateStatus, string> = {
    valid: 'green',
    expiring: 'orange',
    expired: 'default',
    pending: 'blue',
    failed: 'red',
  };
  return colorMap[status] || 'default';
}

/**
 * Get status display text in Chinese
 */
export function getStatusText(status: CertificateStatus): string {
  const textMap: Record<CertificateStatus, string> = {
    valid: '有效',
    expiring: '即将过期',
    expired: '已过期',
    pending: '生成中',
    failed: '生成失败',
  };
  return textMap[status] || '未知';
}

/**
 * Get certificate type color
 */
export function getCertTypeColor(certType: CertificateType | string): string {
  return certType === 'wildcard' ? 'purple' : 'blue';
}

/**
 * Get certificate type display text
 */
export function getCertTypeText(certType: CertificateType | string): string {
  return certType === 'wildcard' ? '泛域名' : '单域名';
}

/**
 * Extract error message from certificate data
 */
export function extractErrorMessage(cert: any): string {
  if (cert.status !== 'failed') return '';
  
  // Try to extract error message from CA field
  if (cert.ca && cert.ca.startsWith('[ERROR]')) {
    return cert.ca.replace('[ERROR]', '').trim();
  }
  
  // Try public_key field (used as error storage)
  if (cert.public_key && (cert.public_key.includes('失败') || cert.public_key.includes('错误'))) {
    return cert.public_key;
  }
  
  return '证书生成失败，请查看错误信息或重新申请';
}

/**
 * Check if CA field contains an error
 */
export function isCaError(ca: string): boolean {
  return ca && ca.startsWith('[ERROR]');
}

/**
 * Format date string to Chinese locale
 */
export function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleString('zh-CN');
}
