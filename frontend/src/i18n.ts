import { createI18n } from 'vue-i18n';

const messages = {
  en: {
    dashboard: {
      title: "Database Dashboard",
      settings: "Settings",
      connect: "Connect",
      disconnect: "Disconnect"
    }
  },
  th: {
    dashboard: {
      title: "แดชบอร์ดฐานข้อมูล",
      settings: "ตั้งค่า",
      connect: "เชื่อมต่อ",
      disconnect: "ตัดการเชื่อมต่อ"
    }
  }
};

const i18n = createI18n({
  legacy: false, // you must set `false`, to use Composition API
  locale: localStorage.getItem('language') || 'en', // set locale
  fallbackLocale: 'en', // set fallback locale
  messages, // set locale messages
});

export default i18n;
