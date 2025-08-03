import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';

import translationJa from './locales/ja/translation.json';

const resources = {
  ja: {
    translation: translationJa,
  },
};

i18n
  .use(initReactI18next) // passes i18n down to react-i18next
  .init({
    resources,
    lng: 'ja', // default language
    fallbackLng: 'ja',

    interpolation: {
      escapeValue: false, // react already escapes by default
    },
  });

export default i18n;
