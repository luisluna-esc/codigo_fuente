'use strict';
(()=>{
  async function exportParticipants(id,format){
    try{
      const actor=window.SGL555?.actual()||currentUser();
      const query=new URLSearchParams({actorId:actor.id,eventId:id,resource:'event-participants-'+format});
      const response=await fetch('/api/music/v610?'+query,{
        headers:{Authorization:'Bearer '+database.settings.serverToken,'X-SGL-Personal':window.SGL555?.personalMode()?'1':'0'},
        cache:'no-store',signal:AbortSignal.timeout(60000)
      });
      if(!response.ok){const p=await response.json().catch(()=>({}));throw Error(p.error||'No se pudo exportar la lista.');}
      const file=await response.blob();
      await saveBlobWithDialog('Participantes_evento.'+format,file,format==='pdf'?'Documento PDF|*.pdf':'Lista CSV|*.csv');
    }catch(err){toast(err.message);}
  }
  window.SGL611Events={exportParticipants};
})();
