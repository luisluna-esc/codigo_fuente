/* Public QR destination is separate from Google/SMTP identity configuration. */
(()=>{
'use strict';
const actual=()=>window.SGL555?.actual?.()||currentUser();
const allowed=()=>['admin','secretaria'].includes(actual()?.role)&&!window.SGL555?.personalMode?.();
const headers=()=>({'Authorization':'Bearer '+database.settings.serverToken});
const endpoint='/api/music/v610';
let generation=1,loading=false;
async function decorate(){
 if(currentPage!=='settings'||!allowed())return;
 const root=document.getElementById('content');if(!root||root.querySelector('#music613_qr_settings'))return;
 const panel=document.createElement('section');panel.id='music613_qr_settings';panel.className='card';
 panel.innerHTML='<h2>Dirección pública para los QR</h2><p>Configura el dominio HTTPS del servidor para que cada QR abra la verificación pública de su certificado. Sin dominio, el QR contiene su código de verificación.</p><form id="music613_qr_form"><label class="field-group">Dirección pública<input class="field" id="music613_qr_url" type="url" placeholder="https://musica.ejemplo.org" maxlength="200"></label><p class="muted" id="music613_qr_effective"></p><button class="btn primary" type="submit" disabled>Guardar dirección</button><p id="music613_qr_status" role="status">Consultando configuración…</p></form><p class="muted">Las copias actuales se actualizan conservando el mismo código de verificación. Al cambiar de dominio, mantén una redirección del dominio anterior para que los certificados impresos sigan verificándose. Deja este campo vacío para usar la dirección configurada en el servidor.</p>';
 root.append(panel);const form=panel.querySelector('form'),status=panel.querySelector('#music613_qr_status'),button=form.querySelector('button');
 try{const {payload}=await v3Api(endpoint+'?'+new URLSearchParams({actorId:actual().id,resource:'qr-settings'}),{headers:headers()});if(!payload.ok)throw Error(payload.error||'No se pudo consultar la dirección.');if(!panel.isConnected)return;generation=payload.restoreGeneration;panel.querySelector('input').value=payload.value||'';effective(panel,payload.effectiveURL);status.textContent='';button.disabled=false}catch(error){status.textContent=error.message}
 form.onsubmit=async event=>{event.preventDefault();if(loading)return;loading=true;button.disabled=true;status.textContent='Guardando…';try{const {payload}=await v3Api(endpoint,{method:'POST',headers:headers(),body:{actorId:actual().id,action:'qr.configure',value:{url:panel.querySelector('input').value.trim()},restoreGeneration:generation}});if(!payload.ok)throw Error(payload.error||'No se pudo guardar.');effective(panel,payload.effectiveURL);status.textContent='Dirección guardada. Se actualizaron los QR de los certificados actuales sin cambiar sus códigos.'}catch(error){status.textContent=error.message}finally{loading=false;button.disabled=false}};
}
function effective(panel,url){panel.querySelector('#music613_qr_effective').textContent=url?'Dirección en uso: '+url:'Sin dirección pública: el QR contendrá el código. Añade el dominio HTTPS para abrir la verificación al escanearlo.'}
const content=setContent;setContent=function(...args){const result=content.apply(this,args);decorate();return result};window.setContent=setContent;
window.MusicSettings613={decorate};
})();
