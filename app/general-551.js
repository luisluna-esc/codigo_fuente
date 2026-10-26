/* SGL 1.0.1: private progress, read-only requirement help and back navigation. */
'use strict';
(() => {
  const personal = user => ['aspirant', 'pastor_aspirant'].includes(user?.role);
  const arrow = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M19 12H5m7-7-7 7 7 7"/></svg>';
  const sheet = document.createElement('link');
  sheet.rel = 'stylesheet'; sheet.href = 'general-551.css?v=6.1.2'; document.head.append(sheet);
  let views = [], lastView = null, restoring = false, scheduled = false, viewActor = '';

  const baseAccess = accessibleAspirants;
  accessibleAspirants = function(user = currentUser()) {
    const people = baseAccess(user) || [];
    return personal(user) ? people.filter(person => person.id === user.id) : people;
  };
  window.accessibleAspirants = accessibleAspirants;
  const baseSelect = selectAspirant;
  selectAspirant = function(id) {
    if (personal(currentUser()) && id !== currentUser().id) return;
    return baseSelect(id);
  };
  window.selectAspirant = selectAspirant;
  const baseJourneyOpen = window.sglJourneyOpen;
  window.sglJourneyOpen = function(program, id = '') {
    if (personal(currentUser()) && id && id !== currentUser().id) return;
    return baseJourneyOpen(program, id);
  };

  function currentView() {
    return {page: currentPage, program: String(currentPage).startsWith('module_') ? currentPage.slice(7) : v3ActiveProgramId,
      owner: activeAspirantId, musicSection: currentPage === 'music_module' ? SGL550.getSection() : '', actor: currentUser()?.id};
  }
  function sameView(a, b) { return !!a && !!b && a.page === b.page && a.program === b.program && a.musicSection === b.musicSection && a.owner === b.owner && a.actor === b.actor; }
  function rememberView(view = lastView || currentView()) {
    if (restoring || !view?.page || view.actor !== currentUser()?.id) return;
    if (!sameView(views.at(-1), view)) views.push({...view});
    views = views.slice(-40);
  }
  async function back() {
    if (document.querySelector('#modal_root .modal-bg')) return closeModal();
    const actor = currentUser(); if (!actor) return;
    let prior;
    while (views.length) { const item = views.pop(); if (item.actor === actor.id && !sameView(item, currentView())) { prior = item; break; } }
    restoring = true;
    try {
      if (!prior) return await go('dashboard', {history: false});
      v3ActiveProgramId = prior.program;
      activeAspirantId = personal(actor) ? actor.id : prior.owner;
      if (prior.page === 'music_module') return await SGL550.open(prior.musicSection || 'home');
      return await go(prior.page, {history: false});
    } finally { restoring = false; lastView = currentView(); }
  }
  const baseGo = go;
  go = function(page, ...args) {
    const from = lastView || currentView(), result = baseGo(page, ...args);
    const finish = value => {
      if (value !== false && !sameView(from, currentView())) rememberView(from);
      lastView = currentView(); decorate(); return value;
    };
    return result?.then ? result.then(finish) : finish(result);
  };
  window.go = go;

  function addBackButton(){document.querySelectorAll('.sgl551-back-bar,#music612_back').forEach(node=>node.remove());}
  function addModalBack(){document.querySelectorAll('.sgl551-modal-back').forEach(node=>node.remove());}
  function showInstructions(id, status, opener) {
    const inModal = !!status.closest('#modal_root');
    const programId = (inModal ? v3ModalProgramId : v3ActiveProgramId) || v3ActiveProgramId;
    const ownerId = personal(currentUser()) ? currentUser().id : (inModal ? modalOwnerId : activeAspirantId);
    const requirement = v3RequirementById(id, programId);
    if (!requirement) return;
    const config = v3ConfigFor(ownerId, requirement, programId) || {};
    const instructions = config.instructions || requirement.observation || 'Desarrolla el requisito con claridad y agrega las evidencias solicitadas.';
    const dialog = document.createElement('dialog');
    dialog.className = 'sgl552-help-dialog';
    dialog.setAttribute('aria-labelledby', 'sgl552_help_title');
    dialog.setAttribute('aria-describedby', 'sgl552_help_instructions');
    dialog.innerHTML = `<div class="sgl552-help-head"><div><h2 id="sgl552_help_title">Cómo llenar</h2><p>${esc(v3Program(programId).menu_label || v3Program(programId).name)} · ${esc(id)}</p></div><button type="button" class="btn ghost" data-close-help aria-label="Cerrar indicaciones">×</button></div><div class="sgl552-help-body"><h3>${esc(requirement.official_text)}</h3><p id="sgl552_help_instructions">${esc(instructions)}</p>${config.example ? `<p><b>Ejemplo:</b> ${esc(config.example)}</p>` : ''}${config.evidence ? `<p><b>Evidencia recomendada:</b> ${esc(config.evidence)}</p>` : ''}${config.note ? `<p><b>Nota:</b> ${esc(config.note)}</p>` : ''}</div><div class="sgl552-help-foot"><button type="button" class="btn primary" data-close-help>Cerrar</button></div>`;
    for (const button of dialog.querySelectorAll('[data-close-help]')) button.onclick = () => dialog.close();
    // Escape must dismiss only this help, preserving any open requirement draft.
    dialog.addEventListener('keydown', event => event.stopPropagation());
    dialog.addEventListener('click', event => event.stopPropagation());
    dialog.addEventListener('close', () => { dialog.remove(); if (opener.isConnected) opener.focus({preventScroll: true}); });
    document.body.append(dialog);
    dialog.showModal();
  }
  function pendingIcons() {
    for (const status of document.querySelectorAll('#content .status.pending, #modal_root .status.pending')) {
      if (status.nextElementSibling?.classList.contains('sgl551-pending-help')) continue;
      const row = status.closest('[data-sgl5059-requirement], [data-sgl5058-requirement]');
      const id = row?.getAttribute('data-sgl5059-requirement') || row?.getAttribute('data-sgl5058-requirement') || (status.closest('#modal_root') ? modalRequirementId : '');
      const help = document.createElement(id ? 'button' : 'span'); help.className = 'sgl551-pending-help';
      help.innerHTML = '<img src="assets/icono-duda.png" width="26" height="26" alt="">';
      if (id) {
        help.type = 'button'; help.title = 'Cómo llenar'; help.setAttribute('aria-label', `Cómo llenar el requisito ${id}`);
        help.onclick = event => { event.preventDefault(); event.stopPropagation(); showInstructions(id, status, help); };
        // Requirement rows also open on Enter/Space; keep these keys on the help button.
        help.onkeydown = event => { if (['Enter', ' '].includes(event.key)) event.stopPropagation(); };
      } else help.setAttribute('aria-label', 'Pendiente');
      status.after(help);
    }
    // Use the same read-only viewer for help in an already open report.
    for (const help of document.querySelectorAll('#modal_root [data-action="help"], #modal_root [data-doc-action="help"]')) {
      if (help.dataset.sgl552Help) continue;
      help.dataset.sgl552Help = 'true';
      help.onclick = event => { event.preventDefault(); event.stopPropagation(); showInstructions(modalRequirementId, help, help); };
      help.onkeydown = event => { if (['Enter', ' '].includes(event.key)) event.stopPropagation(); };
    }
  }
  function annualSendButtons() {
    if (currentPage !== 'requirements' || v3ActiveProgramId !== 'music_ministry') return;
    const owner = targetAspirant(); if (!owner) return;
    const row = document.querySelector('[data-sgl5059-requirement="LM-NUT-1"], [data-sgl5058-requirement="LM-NUT-1"]');
    if (!row) return;
    row.querySelectorAll('.sgl5058-send-review, .sgl5049-send-review, .sgl5059-send-review, .sgl5078-send-document, .sgl5078-send-reading, .sgl551-annual-send').forEach(button => button.remove());
    const progress = v3EnsureProgramProgress(owner.id, 'music_ministry'), status = progress.status['LM-NUT-1'];
    const host = row.querySelector('.sgl5059-status-line') || row.querySelector('.req-actions');
    if (host && personal(currentUser()) && owner.id === currentUser().id && ['in_progress', 'needs_changes'].includes(status) && SGL531.ready(SGL531.context('LM-NUT-1', owner.id))) {
      const button = document.createElement('button'); button.type = 'button'; button.className = 'btn primary small sgl551-annual-send'; button.textContent = 'Enviar a revisión';
      button.onclick = event => { event.stopPropagation(); SGL531.send('LM-NUT-1', owner.id); }; host.append(button);
    }
  }
  function decorate() {
    const actor = currentUser()?.id || '';
    if (viewActor !== actor) { viewActor = actor; views = []; lastView = null; }
    addBackButton(); addModalBack(); pendingIcons(); annualSendButtons();
  }
  const baseContent = setContent;
  setContent = function(...args) { const result = baseContent(...args); decorate(); lastView = currentView(); return result; };
  window.setContent = setContent;
  const baseFilter = filterRequirements;
  filterRequirements = function(...args) { const result = baseFilter(...args); decorate(); return result; };
  window.filterRequirements = filterRequirements;
  const baseModal = showModal;
  showModal = function(...args) { const result = baseModal(...args); addModalBack(); pendingIcons(); return result; };
  window.showModal = showModal;
  const observer = new MutationObserver(() => {
    if (scheduled) return; scheduled = true;
    requestAnimationFrame(() => { scheduled = false; addBackButton(); addModalBack(); pendingIcons(); });
  });
  observer.observe(document.body, {childList: true, subtree: true});
  window.SGL551 = {back, rememberView, resetHistory:()=>{views=[];lastView=null;}};
  decorate(); lastView = currentView();
})();
