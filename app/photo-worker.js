/* SGL 5.0.59 - worker estatico para fotos de Informes. */
'use strict';
const SOF=new Set([192,193,194,195,197,198,199,201,202,203,205,206,207]);
async function headerSize(file){
  const head=new Uint8Array(await file.slice(0,Math.min(file.size,524288)).arrayBuffer());
  if(head.length>=24&&head[0]===137&&head[1]===80&&head[2]===78&&head[3]===71){const v=new DataView(head.buffer,head.byteOffset,head.byteLength);return {width:v.getUint32(16),height:v.getUint32(20)};}
  if(head.length>=4&&head[0]===255&&head[1]===216){let i=2;while(i+9<head.length){if(head[i]!==255){i++;continue;}const m=head[i+1];i+=2;if(m===216||m===217||m===1||(m>=208&&m<=215))continue;if(i+2>head.length)break;const len=(head[i]<<8)|head[i+1];if(len<2||i+len>head.length)break;if(SOF.has(m)&&len>=7)return {height:(head[i+3]<<8)|head[i+4],width:(head[i+5]<<8)|head[i+6]};i+=len;}}
  return null;
}
self.onmessage=async event=>{
  const {id,file,maxBytes,maxDim}=event.data||{};
  try{
    if(!file||!maxBytes||!maxDim)throw new Error('Trabajo de foto incompleto.');
    const meta=await headerSize(file);let targetWidth=meta?.width||0,targetHeight=meta?.height||0;
    if(targetWidth&&targetHeight){const scale=Math.min(1,maxDim/Math.max(targetWidth,targetHeight));targetWidth=Math.max(1,Math.round(targetWidth*scale));targetHeight=Math.max(1,Math.round(targetHeight*scale));}
    const opts={imageOrientation:'from-image',resizeQuality:'medium'};if(targetWidth&&targetHeight){opts.resizeWidth=targetWidth;opts.resizeHeight=targetHeight;}else opts.resizeWidth=maxDim;
    const bitmap=await createImageBitmap(file,opts);
    try{
      let width=bitmap.width,height=bitmap.height;
      const encode=async(quality,w=width,h=height)=>{const canvas=new OffscreenCanvas(w,h),ctx=canvas.getContext('2d',{alpha:false});if(!ctx)throw new Error('Canvas no disponible');ctx.fillStyle='#fff';ctx.fillRect(0,0,w,h);ctx.drawImage(bitmap,0,0,w,h);return canvas.convertToBlob({type:'image/jpeg',quality});};
      let quality=.82,blob=await encode(quality);
          while(blob.size>maxBytes){
            if(quality>.46)quality=Math.max(.46,quality-.12);
            else{if(Math.max(width,height)<=32)throw new Error('No se pudo reducir la foto a 100 KB.');width=Math.max(1,Math.round(width*.8));height=Math.max(1,Math.round(height*.8));quality=.7;}
            blob=await encode(quality,width,height);
          }
      const dataUrl=new FileReaderSync().readAsDataURL(blob);self.postMessage({id,ok:true,dataUrl,size:blob.size,type:'image/jpeg',width,height});
    }finally{bitmap.close?.();}
  }catch(error){self.postMessage({id,ok:false,error:String(error?.message||error||'Error al preparar foto')});}
};
