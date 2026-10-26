'use strict';
window.__MUSICA_WEB__=true;
window.__MUSICA_ONLINE__=true;
const musicFetch=window.fetch.bind(window);
window.fetch=function(input,options){const raw=typeof input==='string'?input:input instanceof URL?input.href:input.url;const url=new URL(raw,location.href);if(url.origin===location.origin&&url.pathname.startsWith('/api/')&&!url.pathname.startsWith('/api/musica/')){url.pathname='/api/musica'+url.pathname;return musicFetch(typeof input==='string'||input instanceof URL?url.href:new Request(url.href,input),options);}return musicFetch(input,options);};
