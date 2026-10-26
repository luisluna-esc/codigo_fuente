/* SGL 5.18 - normalizador PDF sin dependencias externas.
   Reencuadra hasta cinco paginas dentro de hojas Carta conservando el contenido
   PDF original. El archivo se amplía mediante una revision incremental: no se
   rasteriza el texto ni se modifica el contenido fuente. */
(function installSGLPdfNormalizer(global){
  'use strict';
  if(global.SGLPdfNormalizer)return;

  const LETTER_PORTRAIT=[612,792];
  const LETTER_LANDSCAPE=[792,612];
  const WS=/[\x00\x09\x0A\x0C\x0D\x20]/;
  const DELIM=/[\x00\x09\x0A\x0C\x0D\x20\(\)<>{}\[\]\/\%]/;

  function ascii(bytes){
    let out='';
    const step=0x8000;
    for(let i=0;i<bytes.length;i+=step)out+=String.fromCharCode(...bytes.subarray(i,Math.min(bytes.length,i+step)));
    return out;
  }
  function bytes(text){return new TextEncoder().encode(text);}
  function skip(text,pos){
    while(pos<text.length){
      if(WS.test(text[pos])){pos++;continue;}
      if(text[pos]==='%'){while(pos<text.length&&!/[\r\n]/.test(text[pos]))pos++;continue;}
      break;
    }
    return pos;
  }
  function rawNode(raw){return {t:'raw',raw};}
  function numNode(value){return {t:'num',v:Number(value)};}
  function nameNode(value){return {t:'name',v:String(value||'').replace(/^\//,'')};}
  function refNode(obj,gen=0){return {t:'ref',obj:Number(obj),gen:Number(gen)||0};}
  function arrayNode(items=[]){return {t:'array',items};}
  function dictNode(entries=[]){return {t:'dict',entries};}

  function parseValue(text,start=0){
    let pos=skip(text,start);
    if(pos>=text.length)throw new Error('Objeto PDF incompleto.');
    if(text.startsWith('<<',pos)){
      pos+=2;const entries=[];
      while(true){
        pos=skip(text,pos);
        if(text.startsWith('>>',pos)){pos+=2;break;}
        if(text[pos]!=='/')throw new Error('Diccionario PDF no valido.');
        const key=parseName(text,pos);pos=skip(text,key.pos);
        const value=parseValue(text,pos);entries.push({key:key.node.v,value:value.node});pos=value.pos;
      }
      return {node:dictNode(entries),pos};
    }
    if(text[pos]==='['){
      pos++;const items=[];
      while(true){
        pos=skip(text,pos);if(pos>=text.length)throw new Error('Arreglo PDF incompleto.');
        if(text[pos]===']'){pos++;break;}
        const value=parseValue(text,pos);items.push(value.node);pos=value.pos;
      }
      return {node:arrayNode(items),pos};
    }
    if(text[pos]==='/')return parseName(text,pos);
    if(text[pos]==='('){
      const begin=pos;let depth=0,escaped=false;
      while(pos<text.length){
        const ch=text[pos++];
        if(escaped){escaped=false;continue;}
        if(ch==='\\'){escaped=true;continue;}
        if(ch==='(')depth++;
        else if(ch===')'){depth--;if(depth===0)break;}
      }
      return {node:rawNode(text.slice(begin,pos)),pos};
    }
    if(text[pos]==='<'){
      const begin=pos++;while(pos<text.length&&text[pos]!=='>' )pos++;if(pos<text.length)pos++;
      return {node:rawNode(text.slice(begin,pos)),pos};
    }
    const number=text.slice(pos).match(/^[+-]?(?:\d+\.?\d*|\.\d+)/);
    if(number){
      const firstRaw=number[0],first=Number(firstRaw);pos+=firstRaw.length;
      const afterFirst=skip(text,pos),second=text.slice(afterFirst).match(/^\d+/);
      if(Number.isInteger(first)&&first>=0&&second){
        const afterSecond=skip(text,afterFirst+second[0].length);
        if(text[afterSecond]==='R'&&DELIM.test(text[afterSecond+1]||' '))return {node:refNode(first,Number(second[0])),pos:afterSecond+1};
      }
      return {node:{t:'num',v:first,raw:firstRaw},pos};
    }
    const begin=pos;while(pos<text.length&&!DELIM.test(text[pos]))pos++;
    const word=text.slice(begin,pos);
    if(word==='true'||word==='false')return {node:{t:'bool',v:word==='true'},pos};
    if(word==='null')return {node:{t:'null'},pos};
    return {node:rawNode(word),pos};
  }
  function parseName(text,start){
    let pos=start+1;while(pos<text.length&&!DELIM.test(text[pos]))pos++;
    const raw=text.slice(start+1,pos);
    return {node:nameNode(raw.replace(/#([0-9A-Fa-f]{2})/g,(_,hex)=>String.fromCharCode(parseInt(hex,16)))),pos};
  }
  function dictGet(dict,key){
    if(!dict||dict.t!=='dict')return null;
    for(let i=dict.entries.length-1;i>=0;i--)if(dict.entries[i].key===key)return dict.entries[i].value;
    return null;
  }
  function dictSet(dict,key,value){
    if(!dict||dict.t!=='dict')return;
    for(let i=dict.entries.length-1;i>=0;i--)if(dict.entries[i].key===key){dict.entries[i].value=value;return;}
    dict.entries.push({key,value});
  }
  function dictDelete(dict,key){if(dict?.t==='dict')dict.entries=dict.entries.filter(entry=>entry.key!==key);}
  function nodeNumber(node,fallback=0){return node?.t==='num'&&Number.isFinite(node.v)?Number(node.v):fallback;}
  function nodeName(node){return node?.t==='name'?node.v:'';}
  function clone(node){return node==null?node:structuredClone(node);}
  function serialize(node){
    if(!node)return 'null';
    switch(node.t){
      case 'dict':return `<<${node.entries.map(entry=>` /${entry.key} ${serialize(entry.value)}`).join('')} >>`;
      case 'array':return `[${node.items.map(serialize).join(' ')}]`;
      case 'name':return `/${String(node.v||'').replace(/[^!-$&-'*-.0-;=?-Z\\^-z|~]/g,ch=>`#${ch.charCodeAt(0).toString(16).padStart(2,'0').toUpperCase()}`)}`;
      case 'ref':return `${node.obj} ${node.gen||0} R`;
      case 'num':return Number.isInteger(node.v)?String(node.v):String(Number(node.v.toFixed(6)));
      case 'bool':return node.v?'true':'false';
      case 'null':return 'null';
      case 'raw':return node.raw||'';
      default:return 'null';
    }
  }

  async function inflate(data){
    if(typeof DecompressionStream!=='function')throw new Error('Este navegador no puede abrir PDFs con estructura comprimida.');
    const stream=new Blob([data]).stream().pipeThrough(new DecompressionStream('deflate'));
    return new Uint8Array(await new Response(stream).arrayBuffer());
  }
  function filtersFor(dict){
    const node=dictGet(dict,'Filter');
    if(!node)return [];
    if(node.t==='name')return [node.v];
    if(node.t==='array')return node.items.map(nodeName).filter(Boolean);
    return [];
  }
  async function decodeStream(entry){
    let data=entry.streamBytes||new Uint8Array();
    const filters=filtersFor(entry.ast);
    for(const filter of filters){
      if(filter==='FlateDecode'||filter==='Fl')data=await inflate(data);
      else throw new Error(`Filtro PDF no compatible: ${filter}.`);
    }
    return data;
  }

  function scanRegularObjects(sourceBytes,sourceText){
    const map=new Map();let maxObject=0;
    const re=/(^|[\r\n\x00\x09\x0C\x20])(\d+)\s+(\d+)\s+obj\b/g;let match;
    while((match=re.exec(sourceText))){
      const obj=Number(match[2]),gen=Number(match[3]),objStart=match.index+match[1].length;
      const bodyStart=re.lastIndex;
      let endObj=sourceText.indexOf('endobj',bodyStart);if(endObj<0)continue;
      let parsed;
      try{parsed=parseValue(sourceText.slice(bodyStart,endObj),0);}catch{continue;}
      const entry={obj,gen,order:objStart,ast:parsed.node,streamBytes:null,regular:true};
      if(parsed.node?.t==='dict'){
        const local=sourceText.slice(bodyStart,endObj),streamPos=local.indexOf('stream',parsed.pos);
        if(streamPos>=0){
          const absoluteKeyword=bodyStart+streamPos;
          let streamStart=absoluteKeyword+6;
          if(sourceText[streamStart]==='\r'&&sourceText[streamStart+1]==='\n')streamStart+=2;
          else if(sourceText[streamStart]==='\r'||sourceText[streamStart]==='\n')streamStart+=1;
          const endStream=sourceText.indexOf('endstream',streamStart);
          if(endStream>=0){
            const declared=nodeNumber(dictGet(parsed.node,'Length'),-1);let streamEnd=declared>=0&&streamStart+declared<=endStream?streamStart+declared:endStream;if(declared<0){if(sourceText[streamEnd-1]==='\n')streamEnd--;if(sourceText[streamEnd-1]==='\r')streamEnd--;}
            entry.streamBytes=sourceBytes.slice(streamStart,streamEnd);
            endObj=sourceText.indexOf('endobj',endStream+9);if(endObj<0)endObj=endStream+9;
          }
        }
      }
      const previous=map.get(obj);if(!previous||previous.order<=entry.order)map.set(obj,entry);
      maxObject=Math.max(maxObject,obj);re.lastIndex=Math.max(re.lastIndex,endObj+6);
    }
    return {map,maxObject};
  }

  async function expandObjectStreams(objects){
    const containers=[...objects.map.values()].filter(entry=>entry.ast?.t==='dict'&&nodeName(dictGet(entry.ast,'Type'))==='ObjStm'&&entry.streamBytes);
    for(const container of containers){
      const count=nodeNumber(dictGet(container.ast,'N'),0),first=nodeNumber(dictGet(container.ast,'First'),-1);
      if(!count||first<0)continue;
      const decoded=await decodeStream(container),text=ascii(decoded),header=text.slice(0,first).trim().split(/\s+/).map(Number);
      if(header.length<count*2)continue;
      const items=[];for(let i=0;i<count;i++)items.push({obj:header[i*2],offset:header[i*2+1]});
      for(let i=0;i<items.length;i++){
        const item=items[i],start=first+item.offset,end=i+1<items.length?first+items[i+1].offset:text.length;
        if(!Number.isInteger(item.obj)||item.obj<=0||start<first||end<=start)continue;
        let parsed;try{parsed=parseValue(text.slice(start,end),0);}catch{continue;}
        const entry={obj:item.obj,gen:0,order:container.order+((i+1)/(items.length+1)),ast:parsed.node,streamBytes:null,regular:false,container:container.obj};
        const previous=objects.map.get(item.obj);if(!previous||previous.order<=entry.order)objects.map.set(item.obj,entry);
        objects.maxObject=Math.max(objects.maxObject,item.obj);
      }
    }
  }

  function resolved(node,objects){
    if(node?.t!=='ref')return node;
    return objects.map.get(node.obj)?.ast||null;
  }
  function numericArray(node,objects){
    node=resolved(node,objects);if(node?.t!=='array'||node.items.length<4)return null;
    const values=node.items.slice(0,4).map(item=>nodeNumber(resolved(item,objects),NaN));
    return values.every(Number.isFinite)?values:null;
  }
  function rootCatalog(objects){
    let best=null;
    for(const entry of objects.map.values())if(entry.ast?.t==='dict'&&nodeName(dictGet(entry.ast,'Type'))==='Catalog'&&(!best||entry.order>best.order))best=entry;
    return best;
  }
  function pagesFromCatalog(objects,catalog){
    if(!catalog)throw new Error('No se encontro el catalogo del PDF.');
    const pagesRef=dictGet(catalog.ast,'Pages');if(!pagesRef)throw new Error('El PDF no contiene paginas reconocibles.');
    const pages=[];const visited=new Set();
    function walk(nodeOrRef,inherited={}){
      const ref=nodeOrRef?.t==='ref'?nodeOrRef:null,entry=ref?objects.map.get(ref.obj):null,node=ref?entry?.ast:nodeOrRef;
      if(!node||node.t!=='dict')return;
      const visitKey=ref?`${ref.obj}:${ref.gen||0}`:'';if(visitKey){if(visited.has(visitKey))return;visited.add(visitKey);}
      const next={
        mediaBox:dictGet(node,'MediaBox')||inherited.mediaBox||null,
        cropBox:dictGet(node,'CropBox')||inherited.cropBox||null,
        rotate:dictGet(node,'Rotate')||inherited.rotate||null,
        resources:dictGet(node,'Resources')||inherited.resources||null
      };
      const type=nodeName(dictGet(node,'Type'));
      if(type==='Page'||(!dictGet(node,'Kids')&&ref)){
        if(!ref||!entry)throw new Error('Pagina PDF directa no compatible.');
        pages.push({ref,entry,node,inherited:next});return;
      }
      const kids=resolved(dictGet(node,'Kids'),objects);if(kids?.t!=='array')return;
      for(const kid of kids.items)walk(kid,next);
    }
    walk(pagesRef,{});
    return pages;
  }

  function contentItems(page,objects){
    const value=dictGet(page.node,'Contents');if(!value)return [];
    if(value.t==='array')return value.items.map(clone);
    if(value.t==='ref'){
      const target=objects.map.get(value.obj)?.ast;
      if(target?.t==='array')return target.items.map(clone);
    }
    return [clone(value)];
  }
  function pageBox(page,objects){
    return numericArray(page.inherited.cropBox,objects)||numericArray(page.inherited.mediaBox,objects);
  }
  function targetLetter(box){
    const width=Math.abs(box[2]-box[0]),height=Math.abs(box[3]-box[1]);
    return width>height?LETTER_LANDSCAPE:LETTER_PORTRAIT;
  }
  function pageOrder(count,order){
    const fallback=Array.from({length:count},(_,index)=>index);
    if(order==null)return fallback;
    if(!Array.isArray(order)||order.length!==count)throw new Error('El orden de paginas del PDF no es valido.');
    const seen=new Set();
    for(const value of order){if(!Number.isInteger(value)||value<0||value>=count||seen.has(value))throw new Error('El orden de paginas del PDF no es valido.');seen.add(value);}
    return [...order];
  }
  function rootPagesInfo(objects,catalog){
    const ref=dictGet(catalog?.ast,'Pages');
    if(ref?.t!=='ref')throw new Error('El PDF no contiene un arbol de paginas editable.');
    const entry=objects.map.get(ref.obj);
    if(entry?.ast?.t!=='dict')throw new Error('El PDF no contiene un arbol de paginas editable.');
    return {ref,entry};
  }
  function materializePageInheritance(pageDict,page,rootRef){
    if(!dictGet(pageDict,'Resources')&&page.inherited.resources)dictSet(pageDict,'Resources',clone(page.inherited.resources));
    if(!dictGet(pageDict,'Rotate')&&page.inherited.rotate)dictSet(pageDict,'Rotate',clone(page.inherited.rotate));
    dictSet(pageDict,'Parent',clone(rootRef));
  }
  function rootPagesRecord(info,orderedPages){
    const dict=clone(info.entry.ast);
    dictSet(dict,'Kids',arrayNode(orderedPages.map(page=>clone(page.ref))));
    dictSet(dict,'Count',numNode(orderedPages.length));
    return {obj:info.ref.obj,gen:info.ref.gen||info.entry.gen||0,body:serialize(dict)};
  }
  function transformFor(box,target,margin){
    const x0=Math.min(box[0],box[2]),y0=Math.min(box[1],box[3]),sw=Math.abs(box[2]-box[0]),sh=Math.abs(box[3]-box[1]);
    if(!(sw>0&&sh>0))throw new Error('Una pagina del PDF tiene un tamaño invalido.');
    const availableW=Math.max(1,target[0]-margin*2),availableH=Math.max(1,target[1]-margin*2),scale=Math.min(availableW/sw,availableH/sh);
    const drawnW=sw*scale,drawnH=sh*scale,tx=(target[0]-drawnW)/2-x0*scale,ty=(target[1]-drawnH)/2-y0*scale;
    return {scale,tx,ty};
  }
  function setLetterBoxes(dict,target){
    const box=arrayNode([numNode(0),numNode(0),numNode(target[0]),numNode(target[1])]);
    dictSet(dict,'MediaBox',clone(box));dictSet(dict,'CropBox',clone(box));
    for(const key of ['BleedBox','TrimBox','ArtBox'])dictDelete(dict,key);
  }

  function latestStartXref(text){
    const re=/startxref\s+(\d+)\s+%%EOF/g;let match,last=null;while((match=re.exec(text)))last=Number(match[1]);
    if(last==null){const fallback=[...text.matchAll(/startxref\s+(\d+)/g)].pop();last=fallback?Number(fallback[1]):null;}
    return Number.isFinite(last)?last:null;
  }
  function hasEncryption(text){return /\/Encrypt\s+\d+\s+\d+\s+R/.test(text);}
  function concat(chunks,total){const out=new Uint8Array(total);let pos=0;for(const chunk of chunks){out.set(chunk,pos);pos+=chunk.length;}return out;}

  async function normalize(input,options={}){
    const source=input instanceof Uint8Array?input:new Uint8Array(input instanceof ArrayBuffer?input:await input.arrayBuffer());
    if(source.length<8||ascii(source.subarray(0,8)).indexOf('%PDF-')!==0)throw new Error('Selecciona un archivo PDF valido.');
    const sourceText=ascii(source);
    if(hasEncryption(sourceText))throw new Error('El PDF esta protegido o cifrado. Guarda una copia sin contraseña y vuelve a intentarlo.');
    const prev=latestStartXref(sourceText);if(prev==null)throw new Error('El PDF no tiene una estructura de cierre valida.');
    const objects=scanRegularObjects(source,sourceText);await expandObjectStreams(objects);
    const catalog=rootCatalog(objects),pages=pagesFromCatalog(objects,catalog),maxPages=Math.max(1,Number(options.maxPages)||5);
    if(!pages.length)throw new Error('El PDF no contiene paginas reconocibles.');
    if(pages.length>maxPages)throw new Error(`Este requisito admite un maximo de ${maxPages} paginas.`);
    const order=pageOrder(pages.length,options.pageOrder),orderedPages=order.map(index=>pages[index]),rootInfo=rootPagesInfo(objects,catalog);
    const margin=Math.max(0,Math.min(54,Number(options.marginPoints??18)||0));
    let nextObject=objects.maxObject+1;
    const records=[rootPagesRecord(rootInfo,orderedPages)];
    for(const page of orderedPages){
      const box=pageBox(page,objects);if(!box)throw new Error('No se pudo determinar el tamaño de una pagina del PDF.');
      const target=targetLetter(box),matrix=transformFor(box,target,margin),prefix=nextObject++,suffix=nextObject++;
      const pageDict=clone(page.node);materializePageInheritance(pageDict,page,rootInfo.ref);setLetterBoxes(pageDict,target);
      const current=contentItems(page,objects);dictSet(pageDict,'Contents',arrayNode([refNode(prefix,0),...current,refNode(suffix,0)]));
      const prefixBody=`q\n${matrix.scale.toFixed(8)} 0 0 ${matrix.scale.toFixed(8)} ${matrix.tx.toFixed(8)} ${matrix.ty.toFixed(8)} cm\n`;
      const suffixBody='\nQ\n';
      records.push({obj:page.ref.obj,gen:page.ref.gen||page.entry.gen||0,body:serialize(pageDict)});
      records.push({obj:prefix,gen:0,body:`<< /Length ${bytes(prefixBody).length} >>\nstream\n${prefixBody}endstream`});
      records.push({obj:suffix,gen:0,body:`<< /Length ${bytes(suffixBody).length} >>\nstream\n${suffixBody}endstream`});
    }
    records.sort((a,b)=>a.obj-b.obj);
    const chunks=[source],offsets=new Map();let total=source.length;
    const add=text=>{const chunk=bytes(text);chunks.push(chunk);total+=chunk.length;};
    add('\n% SGL 5.18 - paginas normalizadas a tamaño Carta\n');
    for(const record of records){offsets.set(record.obj,{offset:total,gen:record.gen});add(`${record.obj} ${record.gen} obj\n${record.body}\nendobj\n`);}
    const xrefOffset=total;add('xref\n');
    const numbers=[...offsets.keys()].sort((a,b)=>a-b);let cursor=0;
    while(cursor<numbers.length){let end=cursor+1;while(end<numbers.length&&numbers[end]===numbers[end-1]+1)end++;
      const group=numbers.slice(cursor,end);add(`${group[0]} ${group.length}\n`);
      for(const obj of group){const item=offsets.get(obj);add(`${String(item.offset).padStart(10,'0')} ${String(item.gen||0).padStart(5,'0')} n \n`);}
      cursor=end;
    }
    const size=Math.max(nextObject,objects.maxObject+1),rootRef=refNode(catalog.obj,catalog.gen||0);
    add(`trailer\n<< /Size ${size} /Root ${serialize(rootRef)} /Prev ${prev} >>\nstartxref\n${xrefOffset}\n%%EOF\n`);
    const output=concat(chunks,total);
    return {blob:new Blob([output],{type:'application/pdf'}),bytes:output,pageCount:pages.length,normalized:true};
  }

  async function reorder(input,order,options={}){
    const source=input instanceof Uint8Array?input:new Uint8Array(input instanceof ArrayBuffer?input:await input.arrayBuffer());
    if(source.length<8||ascii(source.subarray(0,8)).indexOf('%PDF-')!==0)throw new Error('Selecciona un archivo PDF valido.');
    const sourceText=ascii(source);
    if(hasEncryption(sourceText))throw new Error('El PDF esta protegido o cifrado. Guarda una copia sin contraseña y vuelve a intentarlo.');
    const prev=latestStartXref(sourceText);if(prev==null)throw new Error('El PDF no tiene una estructura de cierre valida.');
    const objects=scanRegularObjects(source,sourceText);await expandObjectStreams(objects);
    const catalog=rootCatalog(objects),pages=pagesFromCatalog(objects,catalog),maxPages=Math.max(1,Number(options.maxPages)||5);
    if(!pages.length)throw new Error('El PDF no contiene paginas reconocibles.');
    if(pages.length>maxPages)throw new Error(`Este requisito admite un maximo de ${maxPages} paginas.`);
    const normalized=pageOrder(pages.length,order),orderedPages=normalized.map(index=>pages[index]),rootInfo=rootPagesInfo(objects,catalog),records=[rootPagesRecord(rootInfo,orderedPages)];
    for(const page of orderedPages){
      const pageDict=clone(page.node);materializePageInheritance(pageDict,page,rootInfo.ref);
      records.push({obj:page.ref.obj,gen:page.ref.gen||page.entry.gen||0,body:serialize(pageDict)});
    }
    records.sort((a,b)=>a.obj-b.obj);
    const chunks=[source],offsets=new Map();let total=source.length;
    const add=text=>{const chunk=bytes(text);chunks.push(chunk);total+=chunk.length;};
    add('\n% SGL - orden de paginas actualizado\n');
    for(const record of records){offsets.set(record.obj,{offset:total,gen:record.gen});add(`${record.obj} ${record.gen} obj\n${record.body}\nendobj\n`);}
    const xrefOffset=total;add('xref\n');
    const numbers=[...offsets.keys()].sort((a,b)=>a-b);let cursor=0;
    while(cursor<numbers.length){let end=cursor+1;while(end<numbers.length&&numbers[end]===numbers[end-1]+1)end++;const group=numbers.slice(cursor,end);add(`${group[0]} ${group.length}\n`);for(const obj of group){const item=offsets.get(obj);add(`${String(item.offset).padStart(10,'0')} ${String(item.gen||0).padStart(5,'0')} n \n`);}cursor=end;}
    const size=Math.max(objects.maxObject+1,...numbers.map(value=>value+1)),rootRef=refNode(catalog.obj,catalog.gen||0);
    add(`trailer\n<< /Size ${size} /Root ${serialize(rootRef)} /Prev ${prev} >>\nstartxref\n${xrefOffset}\n%%EOF\n`);
    const output=concat(chunks,total);
    return {blob:new Blob([output],{type:'application/pdf'}),bytes:output,pageCount:pages.length,pageOrder:normalized,reordered:true};
  }

  async function inspect(input,options={}){
    const source=input instanceof Uint8Array?input:new Uint8Array(input instanceof ArrayBuffer?input:await input.arrayBuffer()),text=ascii(source);
    const objects=scanRegularObjects(source,text);await expandObjectStreams(objects);const catalog=rootCatalog(objects),pages=pagesFromCatalog(objects,catalog);
    const maxPages=Math.max(1,Number(options.maxPages)||5);return {pageCount:pages.length,withinLimit:pages.length>0&&pages.length<=maxPages,encrypted:hasEncryption(text)};
  }

  async function merge(inputs){
    if(!inputs.length)throw new Error('No hay documentos para reunir.');
    const records=[],kids=[];let next=3;
    function remap(node,ids){if(!node)return node;const n=clone(node);if(n.t==='ref'){if(!ids.has(n.obj))throw new Error('Referencia PDF incompleta.');return refNode(ids.get(n.obj),0)}if(n.t==='array')n.items=n.items.map(x=>remap(x,ids));if(n.t==='dict')n.entries=n.entries.map(e=>({...e,value:remap(e.value,ids)}));return n}
    for(const input of inputs){
      const source=new Uint8Array(input instanceof Uint8Array?input:await input.arrayBuffer()),text=ascii(source);
      if(hasEncryption(text))throw new Error('No se pueden reunir PDF cifrados.');
      const objects=scanRegularObjects(source,text);await expandObjectStreams(objects);const pages=pagesFromCatalog(objects,rootCatalog(objects));if(!pages.length)throw new Error('Uno de los PDF no contiene páginas.');
      const ids=new Map([...objects.map.keys()].map(id=>[id,next++])),pageMap=new Map(pages.map(p=>[p.ref.obj,p]));kids.push(...pages.map(p=>refNode(ids.get(p.ref.obj))));
      for(const entry of objects.map.values()){
        let ast=clone(entry.ast);const page=pageMap.get(entry.obj);
        if(page){for(const [key,field] of [['MediaBox','mediaBox'],['CropBox','cropBox'],['Resources','resources'],['Rotate','rotate']])if(!dictGet(ast,key)&&page.inherited[field])dictSet(ast,key,clone(page.inherited[field]));}
        ast=remap(ast,ids);if(page){dictSet(ast,'Parent',refNode(2));}
        if(entry.streamBytes)dictSet(ast,'Length',numNode(entry.streamBytes.length));
        records.push({id:ids.get(entry.obj),ast,stream:entry.streamBytes});
      }
    }
    records.push({id:1,ast:dictNode([{key:'Type',value:nameNode('Catalog')},{key:'Pages',value:refNode(2)}])},{id:2,ast:dictNode([{key:'Type',value:nameNode('Pages')},{key:'Count',value:numNode(kids.length)},{key:'Kids',value:arrayNode(kids)}])});
    const chunks=[],offsets=[0];let total=0;const add=v=>{const b=typeof v==='string'?bytes(v):v;chunks.push(b);total+=b.length};add('%PDF-1.7\n');
    for(const r of records){offsets[r.id]=total;add(`${r.id} 0 obj\n${serialize(r.ast)}`);if(r.stream){add('\nstream\n');add(r.stream);add('\nendstream')}add('\nendobj\n')}
    const xref=total;add(`xref\n0 ${next}\n0000000000 65535 f \n`);for(let i=1;i<next;i++)add(`${String(offsets[i]).padStart(10,'0')} 00000 n \n`);add(`trailer\n<< /Size ${next} /Root 1 0 R >>\nstartxref\n${xref}\n%%EOF\n`);return new Blob(chunks,{type:'application/pdf'});
  }

  global.SGLPdfNormalizer={merge,normalize,reorder,inspect,LETTER_PORTRAIT:[...LETTER_PORTRAIT],LETTER_LANDSCAPE:[...LETTER_LANDSCAPE]};
})(typeof window!=='undefined'?window:globalThis);
