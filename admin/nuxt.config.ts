import tailwindcss from "@tailwindcss/vite";

export default defineNuxtConfig({
  devtools: { enabled: true },
  components: [{ path: '~/components', extensions: ['vue'] }],
  css: ['~/assets/css/main.css'],
  devServer: {
    host: '0.0.0.0',
    port: 3000,
  },
  runtimeConfig: {
    public: {
      apiBase: 'https://srvr.novdaunion.uz',
    },
  },
  vite: {
    plugins: [tailwindcss()]
  }
})
