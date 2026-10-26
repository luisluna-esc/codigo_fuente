/* Login reminder: assigned applicants only, fifteen seconds, retained in the bell. */
'use strict';
(()=>{
  let shownSession='', timer=0, pendingKeys=new Set(), ready=false;
  let pendingCount=0;
  const enabled=()=>!['admin','departamental','secretaria','coordinador_musica','administrativo_sgm'].includes((window.SGL555?.actual()||currentUser())?.role);
  const key=row=>[row.ownerId,row.requirementId,row.submittedAt].join('|');
  function dismiss(){
    clearTimeout(timer);
    document.getElementById('music611_review_alert')?.remove();
  }
  async function review(report=false){
    dismiss();
    if(await closeModal()===false)return;
    if(window.SGL555?.personalMode()){
      if(!await SGL555.switchView('management'))return;
    }
    if(report)return SGL610.reviewReport();
    go('reviews');
  }
  function update(view){
    document.getElementById('music610_followup')?.remove();
    if(!view||!currentUser())return;
    ready=true;
    pendingCount=Number(view.pendingCount)||0;
    refreshIcon();
    pendingKeys=new Set((view.pendingRows||[]).map(key));
    if(!pendingCount||!enabled()||view.showReminder===false){dismiss();return;}
    const actor=window.SGL555?.actual()||currentUser();
    const session=actor.id+'|'+(database.settings?.serverToken||'');
    if(shownSession===session)return;
    shownSession=session;
    const box=document.createElement('aside');
    box.id='music611_review_alert';
    box.className='card music611-review-alert';
    box.setAttribute('role','alert');
    box.innerHTML=`<button class="music611-alert-close" type="button" aria-label="Cerrar aviso" onclick="SGL611Alerts.dismiss()">×</button>
      <h2>Tienes requisitos pendientes de revisión</h2>
      <p>Hay ${Number(view.pendingCount)} requisito${view.pendingCount===1?'':'s'} de tus aspirantes que llevan 3 días o más esperando tu revisión.</p>
      <p>Al cumplir 5 días desde su envío se notificará al equipo de gestión para dar seguimiento.</p>
      <button type="button" class="btn primary" onclick="SGL611Alerts.review()">Revisar requisitos</button>
      <small>Este aviso se cierra en 15 segundos. Puedes consultarlo en la campanita.</small>`;
    document.body.append(box);
    timer=setTimeout(dismiss,15000);
    Promise.resolve(window.v4FetchNotifications?.(false,false)).catch(()=>{});
  }
  function visible(item){
    if(item?.type!=='revision610'||!String(item.sourceOperationId||'').startsWith('revision610:recordatorio:'))return true;
    if(!ready)return false;
    const parts=String(item.sourceOperationId).split(':');
    return pendingKeys.has([item.ownerId||parts[2],item.requirementId||parts[3],parts.slice(4).join(':')].join('|'));
  }
  const previous=sgl504VisibleNotifications;
  sgl504VisibleNotifications=function(items){return previous.call(this,items).filter(visible);};
  window.sgl504VisibleNotifications=sgl504VisibleNotifications;
  const kind=v4NotificationKind;
  v4NotificationKind=function(item){return item?.type==='revision610'?'Seguimiento de revisión':kind(item);};
  window.v4NotificationKind=v4NotificationKind;
  const show=v4ShowNotifications;
  v4ShowNotifications=function(fromLogin=false){
    // These reminders use the timed alert; the bell remains available on click.
    if(fromLogin&&(database.syncNotifications||[]).some(n=>n.type==='revision610'&&visible(n)))return;
    return show.apply(this,arguments);
  };
  window.v4ShowNotifications=v4ShowNotifications;
  const actions=v4219NotificationActions;
  v4219NotificationActions=function(item){
    const html=actions(item);
    if(item?.type!=='revision610')return html;
    const report=String(item.sourceOperationId||'').startsWith('revision610:reporte:');
    return html+`<button type="button" class="btn ghost small" onclick="SGL611Alerts.review(${report})">${report?'Ver seguimiento':'Ver revisiones'}</button>`;
  };
  window.v4219NotificationActions=v4219NotificationActions;
  async function message(){
    if(!pendingCount||!enabled()||await closeModal()===false)return;
    showModal(`<div class="modal small music619-review-message"><div class="modal-head"><h2>Pendientes de revisión</h2><button class="modal-close" aria-label="Cerrar aviso" onclick="closeModal()">×</button></div><div class="modal-body"><p>Tienes ${pendingCount} requisito${pendingCount===1?'':'s'} de tus aspirantes que llevan 3 días o más esperando tu revisión.</p><p>Al cumplir 5 días desde su envío se notificará al equipo de gestión para dar seguimiento.</p></div><div class="modal-foot"><button class="btn ghost" onclick="closeModal()">Cerrar</button><button class="btn primary" onclick="SGL611Alerts.review()">Revisar requisitos</button></div></div>`);
  }
  function refreshIcon(){
    const bar=document.querySelector('.topbar');if(!bar)return;
    let button=document.getElementById('music615_review');
    if(!pendingCount||!enabled()){button?.remove();return;}
    if(!button){
      button=document.createElement('button');button.id='music615_review';button.type='button';button.className='btn ghost music615-top-icon';button.onclick=()=>message();
      button.innerHTML='<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="m12 3 10 18H2L12 3Z M12 9v5 M12 17h.01"/></svg><span class="count"></span>';
    }
    const label=pendingCount+' requisito'+(pendingCount===1?'':'s')+' pendiente'+(pendingCount===1?'':'s')+' de revisión';
    button.title=label;button.setAttribute('aria-label',label);button.querySelector('.count').textContent=String(pendingCount);
    const sibling=bar.querySelector('#music615_birthday')||bar.querySelector('#notification_pill');
    if(sibling&&sibling.nextElementSibling!==button)sibling.after(button);else if(!button.isConnected)bar.append(button);
  }
  window.SGL611Alerts={update,dismiss,review,refreshIcon};
})();
