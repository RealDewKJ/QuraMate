import { useColorMode, useDark } from '@vueuse/core';

export const colorMode = useColorMode({
  emitAuto: true,
});

export const isDarkTheme = useDark();
