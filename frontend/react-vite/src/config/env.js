// Environment configuration
export const config = {
  API_URL: import.meta.env.VITE_API_URL || 'http://localhost:8090',
  PUBLIC_SITE_URL: import.meta.env.VITE_PUBLIC_SITE_URL || 'http://localhost:3000',
  TELEGRAM_BOT_LINK: import.meta.env.VITE_TELEGRAM_BOT_LINK || 'https://t.me/your_telegram_bot',
  TELEGRAM_CHANNEL_LINK: import.meta.env.VITE_TELEGRAM_CHANNEL_LINK || 'https://t.me/your_telegram_channel',
  TELEGRAM_SUPPORT_LINK: import.meta.env.VITE_TELEGRAM_SUPPORT_LINK || 'https://t.me/your_support_link',
  INTERNAL_TOKEN: import.meta.env.VITE_INTERNAL_TOKEN || '',
  FRONTEND_SECRET: import.meta.env.VITE_FRONTEND_SECRET || '',
  DEV_MODE: import.meta.env.VITE_DEV_MODE === 'true',
};

export default config;
