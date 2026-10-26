/* Editing on the sheet uses the same draft and rich-text history as the side panel. */
window.MusicInline621=({field,change,range,html,display,fit})=>{
 let active=null;
 function offset(node){const value=node.cloneNode(true);for(const br of value.querySelectorAll('br'))br.replaceWith('\n');for(const block of value.querySelectorAll('div,p'))block.prepend('\n');return value.textContent.length}
 function capture(){if(!active)return;const s=getSelection();if(!s.rangeCount||!active.node.contains(s.anchorNode)||!active.node.contains(s.focusNode))return;const r=s.getRangeAt(0),before=r.cloneRange();before.selectNodeContents(active.node);before.setEnd(r.startContainer,r.startOffset);range({start:offset(before.cloneContents()),end:offset(before.cloneContents())+offset(r.cloneContents())},active.ref)}
 function finish(){
  if(!active)return;capture();const {node,ref}=active;active=null;node.removeAttribute('contenteditable');node.classList.remove('music621-editing');node.setAttribute('role','button');node.removeAttribute('aria-multiline');node.removeEventListener('input',input);node.removeEventListener('blur',finish);node.removeEventListener('keydown',key);document.removeEventListener('selectionchange',capture);node.innerHTML=display(field(ref),field(ref).text);fit();
 }
 function value(node){return node.innerText.replace(/\r\n?/g,'\n')}
 function input(){if(!active)return;change(value(active.node));capture()}
 function key(event){if(event.key==='Escape'||event.key==='Enter'&&(event.ctrlKey||event.metaKey)){event.preventDefault();event.stopPropagation();const node=active.node;finish();node.focus({preventScroll:true})}}
 function start(ref,node){
  finish();active={ref,node};node.innerHTML=html(field(ref),field(ref).text);node.contentEditable='plaintext-only';node.classList.add('music621-editing');node.setAttribute('role','textbox');node.setAttribute('aria-multiline','true');node.addEventListener('input',input);node.addEventListener('blur',finish);node.addEventListener('keydown',key);document.addEventListener('selectionchange',capture);node.focus({preventScroll:true});const r=document.createRange();r.selectNodeContents(node);r.collapse(false);getSelection().removeAllRanges();getSelection().addRange(r);capture();
 }
 return {start,finish,capture,active:()=>!!active};
};
