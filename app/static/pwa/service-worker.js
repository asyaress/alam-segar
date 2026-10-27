const CACHE_VERSION = 'alam-segar-shell-v3';
const SHELL_ASSETS = [
  '/pwa/offline.html',
  '/pwa/icon.svg',
  '/pwa/icon-maskable.svg',
  '/assets/admin/css/style.css'
];

self.addEventListener('install', event => {
  event.waitUntil(caches.open(CACHE_VERSION).then(cache => cache.addAll(SHELL_ASSETS)));
  self.skipWaiting();
});

self.addEventListener('activate', event => {
  event.waitUntil(caches.keys().then(keys => Promise.all(
    keys.filter(key => key !== CACHE_VERSION).map(key => caches.delete(key))
  )));
  self.clients.claim();
});

self.addEventListener('fetch', event => {
  const request = event.request;
  if (request.method !== 'GET') return;
  const url = new URL(request.url);
  if (url.origin !== self.location.origin) return;

  // Data portal, dokumen privat, kuitansi, dan halaman autentikasi tidak
  // disimpan. Navigasi selalu mencoba server agar data finansial tetap aktual.
  if (request.mode === 'navigate') {
    event.respondWith(fetch(request).catch(() => caches.match('/pwa/offline.html')));
    return;
  }
  const safeStatic = url.pathname.startsWith('/assets/admin/') || url.pathname.startsWith('/pwa/');
  if (!safeStatic) return;
  event.respondWith(caches.match(request).then(cached => cached || fetch(request).then(response => {
    if (response.ok) {
      const copy = response.clone();
      caches.open(CACHE_VERSION).then(cache => cache.put(request, copy));
    }
    return response;
  })));
});
