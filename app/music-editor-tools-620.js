/* Text boxes use a private editor clipboard; text inputs retain native copy/paste. */
window.MusicEditorTools620=({current,model,selected,history,insert,status})=>{
let clipboard=null,colorFrame=0;
const normalize=value=>{let v=String(value||'').trim().replace(/^#/,'');if(/^[0-9a-f]{3}$/i.test(v))v=[...v].map(c=>c+c).join('');return /^[0-9a-f]{6}$/i.test(v)?'#'+v.toUpperCase():null};
function copy(){if(selected().kind!=='field'||!current())return false;clipboard=structuredClone(current());toast('Cuadro de texto copiado.');return true}
function paste(){if(!clipboard)return false;return insert(structuredClone(clipboard))}
function duplicate(){if(selected().kind!=='field'||!current())return false;return insert(structuredClone(current()))}
function color(value){
 const next=normalize(value),f=current();if(!next||selected().kind!=='field'||!f)return false;
 if(f.color===next&&!(f.runs618||[]).some(r=>r.color!==next))return true;
 history('color:'+selected().index);f.color=next;for(const r of f.runs618||[])r.color=next;
 const draft=model(),index=selected().index;
 cancelAnimationFrame(colorFrame);colorFrame=requestAnimationFrame(()=>{
  if(model()!==draft||selected().kind!=='field'||selected().index!==index||current()!==f)return;
  const box=document.querySelector(`#music592_stage [data-kind=field][data-item="${index}"]`);
  if(box){box.style.color=f.color;for(const span of box.querySelectorAll('span'))span.style.color=f.color}
  const picker=document.getElementById('music620_color_picker'),hex=document.getElementById('music620_color_hex');
  if(picker)picker.value=f.color;if(hex&&document.activeElement!==hex){hex.value=f.color;hex.setCustomValidity('');hex.setAttribute('aria-invalid','false')}
 });
 return true;
}
function hex(input,commit=false){const next=normalize(input.value);input.setAttribute('aria-invalid',String(!next));input.setCustomValidity(next?'':'Escribe un color hexadecimal, por ejemplo #173C76.');if(!next){if(commit)status('Revisa el color hexadecimal. Puedes pegarlo con o sin #.',true);return false}color(next);if(commit)input.value=next;return true}
function validColor(){const input=document.getElementById('music620_color_hex');if(!input)return true;const value=normalize(input.value);if(value){input.value=value;input.setCustomValidity('');input.setAttribute('aria-invalid','false');return true}hex(input,true);input.focus();return false}
function controls(f){return `<label class="music620-color">Color · HEX<input id="music620_color_hex" type="text" inputmode="text" aria-label="Color hexadecimal" value="${normalize(f.color)||'#173C76'}" maxlength="12" spellcheck="false" autocomplete="off" oninput="SGL592.colorHex(this)" onchange="SGL592.colorHex(this,true)" placeholder="#173C76"><input id="music620_color_picker" type="color" aria-label="Selector de color" value="${normalize(f.color)||'#173C76'}" oninput="SGL592.colorValue(this.value)"></label>`}
return {copy,paste,duplicate,color,hex,validColor,controls};
};
