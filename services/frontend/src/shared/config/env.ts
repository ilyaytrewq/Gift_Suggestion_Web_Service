const configuredApiBaseUrl = import.meta.env.VITE_API_BASE_URL;

export const API_BASE_URL =
  configuredApiBaseUrl === undefined
    ? 'http://localhost:8080'
    : configuredApiBaseUrl.trim().replace(/\/$/, '');
