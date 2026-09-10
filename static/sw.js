// Minimal service worker for the kiosk PWA.
//
// It caches nothing on purpose: documents and signature state must always come
// from the server. It exists only because Chrome requires a registered service
// worker with a fetch handler before it treats the app as installable, and an
// installed app is what gives the home-screen shortcut fullscreen display.

const VERSION = 'v1';

self.addEventListener('install', () => {
  self.skipWaiting();
});

self.addEventListener('activate', (event) => {
  event.waitUntil(self.clients.claim());
});

self.addEventListener('fetch', () => {
  // No-op: every request falls through to the network.
});
