/* Personal settings are available to every role; management actions keep permissions. */
'use strict';
(()=>{
  const actor=()=>window.SGL555?.actual()||currentUser();
  const key=()=> 'musica_preferences_615_'+actor()?.id;
  function preferences(){try{return JSON.parse(localStorage.getItem(key())||'{}');}catch{return {};}}
  function apply(){
    const p=preferences();document.documentElement.classList.toggle('music615-large-text',p.text==='large');
    document.documentElement.classList.toggle('sgm603-reduce-motion',SGL602.permit('ui.reduceMotion')||!!p.reduceMotion);
  }
  function decorate(){
    if(!currentUser())return;apply();
    if(currentPage!=='settings')return;
    const root=document.getElementById('content');if(!root)return;
    root.querySelector('#music_online_preferences')?.remove();
    if(root.querySelector('#music615_preferences'))return;
    const user=actor(),p=preferences(),personal=['aspirant','pastor_aspirant'].includes(currentUser()?.role);
    const panel=document.createElement('section');panel.id='music615_preferences';panel.className='card music615-personal-options';
    panel.innerHTML=`<h2>Mis preferencias</h2><form onsubmit="MusicPreferences615.save(event)"><div class="music615-preferences-grid">
      <label class="field-group">Apariencia<select class="field" name="appearance">${[['light','Día'],['blue','Azul'],['dark','Noche']].map(([id,name])=>`<option value="${id}" ${document.documentElement.dataset.sgl545Mode===id?'selected':''}>${name}</option>`).join('')}</select></label>
      <label class="field-group">Tamaño del texto de la interfaz<select class="field" name="text"><option value="normal">Normal</option><option value="large" ${p.text==='large'?'selected':''}>Grande</option></select></label>
      <label class="field-group">Letra de mis informes<select class="field" name="font">${['Arial','Times New Roman'].map(font=>`<option ${user.reportFont===font?'selected':''}>${font}</option>`).join('')}</select></label>
      ${user.role==='admin'&&!personal?'<label class="field-group">Logo de Música (opcional)<input class="field" name="logo" type="file" accept="image/*"></label>':''}
      </div><label class="choice"><input type="checkbox" name="reduceMotion" ${p.reduceMotion||SGL602.permit('ui.reduceMotion')?'checked':''} ${SGL602.permit('ui.reduceMotion')?'disabled':''}> Reducir animaciones</label>
      <div class="actions"><button class="btn primary" type="submit">Guardar preferencias</button></div><p id="music615_preferences_status" role="status"></p></form>
      <div class="actions"><button class="btn ghost" type="button" onclick="go('profile')">Mi perfil y seguridad</button><button class="btn ghost" type="button" onclick="v4ShowNotifications()">Mis avisos</button>${SGL602.permit('birthdays.calendar')?'<button class="btn ghost" type="button" onclick="sgl603ShowBirthdays()">Ver cumpleaños</button>':''}<a class="btn ghost" href="manual.html" target="_blank" rel="noopener">Ayuda y manual</a></div>
      <p class="muted">La contraseña y el correo se gestionan desde Mi Perfil. Los avisos de revisión se muestran únicamente a sus responsables.</p>`;
    const head=root.querySelector('.page-head');if(head)head.after(panel);else root.prepend(panel);
    if(!personal&&({administrativo_sgm:'secretaria'}[user.role]||user.role)==='administrativo'&&!root.querySelector('#music615_work_options')){
      const work=document.createElement('section');work.id='music615_work_options';work.className='card music615-personal-options';
      work.innerHTML=`<h2>Mis funciones de gestión</h2><p>Accede a tus aspirantes y a las funciones autorizadas.</p><div class="actions"><button class="btn ghost" type="button" onclick="go('aspirants')">Mis aspirantes</button><button class="btn ghost" type="button" onclick="go('music_ministries')">Ministerios y vínculos</button></div><p class="muted">La posibilidad de registrar ministerios depende de Permisos y funciones. Para corregir o eliminar un registro, contacta al equipo de gestión.</p>`;
      root.append(work);
    }
  }
  async function save(event){
    event.preventDefault();const form=event.target,button=form.querySelector('button[type=submit]'),status=form.querySelector('[role=status]');button.disabled=true;status.textContent='Guardando…';
    try{
      const fields=new FormData(form),user=userById(actor().id);user.reportFont=fields.get('font');
      const logo=fields.get('logo');if(user.role==='admin'&&logo?.size)database.settings.musicAppLogo=await imageToTransparentDataUrl(logo,1200);
      await persist(true,{action:'Preferencias guardadas',entity:user.id,detail:'Letra de informes'});
      localStorage.setItem(key(),JSON.stringify({text:fields.get('text'),reduceMotion:fields.has('reduceMotion')}));
      sgl548SetTheme(fields.get('appearance'));apply();window.MusicShell615?.decorate();status.textContent='Preferencias guardadas.';
    }catch(error){status.textContent=error.message;}finally{button.disabled=false;}
  }
  const content=setContent;setContent=function(...args){const out=content.apply(this,args);decorate();return out;};window.setContent=setContent;
  const settings=renderSettings;renderSettings=function(...args){const out=settings.apply(this,args);decorate();return out;};window.renderSettings=renderSettings;
  window.MusicPreferences615={decorate,save,apply};
})();
