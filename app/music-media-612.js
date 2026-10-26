/* New photo uploads only. Certificate artwork and issued PDFs keep their originals. */
'use strict';
(() => {
  const root = globalThis, inWorker = typeof document === 'undefined';
  const presets = {
    report: {maxBytes: 100000, maxDimension: 1600, quality: .86, minQuality: .66},
    ministry: {maxBytes: 220 * 1024, maxDimension: 1440, quality: .86, minQuality: .70},
    profile: {maxBytes: 120 * 1024, maxDimension: 960, quality: .86, minQuality: .70},
    document: {maxBytes: 600 * 1024, maxDimension: 2600, quality: .92, minQuality: .80}
  };
  const MAX_INPUT = 20 * 1024 * 1024, MAX_PIXELS = 64 * 1000 * 1000;
  const toURL = blob => {
    if (inWorker) return Promise.resolve(new FileReaderSync().readAsDataURL(blob));
    return new Promise((resolve, reject) => {
      const reader = new FileReader(); reader.onload = () => resolve(reader.result);
      reader.onerror = () => reject(Error('No se pudo leer la imagen.')); reader.readAsDataURL(blob);
    });
  };
  function exifOrientation(bytes, start, length) {
    try {
      if (length < 14 || String.fromCharCode(...bytes.subarray(start, start + 4)) !== 'Exif') return 1;
      const view = new DataView(bytes.buffer, bytes.byteOffset + start + 6, length - 6);
      const little = view.getUint16(0) === 0x4949;
      if (view.getUint16(2, little) !== 42) return 1;
      const offset = view.getUint32(4, little), count = view.getUint16(offset, little);
      for (let i = 0; i < Math.min(count, 256); i++) {
        const entry = offset + 2 + i * 12;
        if (view.getUint16(entry, little) === 0x112) return view.getUint16(entry + 8, little);
      }
    } catch {}
    return 1;
  }
  async function header(file) {
    const bytes = new Uint8Array(await file.slice(0, 524288).arrayBuffer());
    if (bytes.length >= 24 && bytes[0] === 137 && bytes[1] === 80 && bytes[2] === 78 && bytes[3] === 71) {
      const view = new DataView(bytes.buffer);
      return {width: view.getUint32(16), height: view.getUint32(20), type: 'image/png'};
    }
    if (bytes[0] !== 255 || bytes[1] !== 216) return null;
    const sof = new Set([192, 193, 194, 195, 197, 198, 199, 201, 202, 203, 205, 206, 207]);
    let i = 2, size = null, orientation = 1;
    while (i + 4 <= bytes.length) {
      if (bytes[i++] !== 255) continue;
      let marker = bytes[i++]; while (marker === 255) marker = bytes[i++];
      if (marker === 218 || marker === 217) break;
      if (marker === 216 || marker === 1 || (marker >= 208 && marker <= 215)) continue;
      const length = (bytes[i] << 8) | bytes[i + 1];
      if (length < 2 || i + length > bytes.length) break;
      if (marker === 225) orientation = exifOrientation(bytes, i + 2, length - 2);
      if (sof.has(marker) && length >= 7) size = {height: (bytes[i + 3] << 8) | bytes[i + 4], width: (bytes[i + 5] << 8) | bytes[i + 6], type: 'image/jpeg'};
      i += length;
    }
    if (size) size.orientation = orientation;
    if (size && orientation >= 5 && orientation <= 8) [size.width, size.height] = [size.height, size.width];
    return size;
  }
  function dimensions(width, height, max) {
    const scale = Math.min(1, max / Math.max(width, height));
    return {width: Math.max(1, Math.round(width * scale)), height: Math.max(1, Math.round(height * scale))};
  }
  async function decode(file, meta, max) {
    if (typeof createImageBitmap === 'function') {
      const options = {imageOrientation: 'from-image', resizeQuality: 'high'};
      if (meta) { const size = dimensions(meta.width, meta.height, max); options.resizeWidth = size.width; options.resizeHeight = size.height; }
      try {
        const image = await createImageBitmap(file, options);
        if (!image.width || !image.height || image.width * image.height > MAX_PIXELS) { image.close(); throw Error('La imagen tiene demasiados píxeles.'); }
        return {image, ...dimensions(image.width, image.height, max), close: () => image.close()};
      } catch (error) { if (inWorker) throw error; }
    }
    if (inWorker) throw Error('Procesador de imágenes no disponible.');
    if ((!meta && file.size > 8 * 1024 * 1024) || (meta && meta.width * meta.height > 20 * 1000 * 1000)) throw Error('La imagen es demasiado grande para este navegador. Usa una copia de menor resolución.');
    const url = URL.createObjectURL(file), image = new Image();
    try {
      await new Promise((resolve, reject) => { image.onload = resolve; image.onerror = () => reject(Error('No se pudo abrir la imagen. Usa PNG, JPG o WebP.')); image.src = url; });
      if (!image.naturalWidth || image.naturalWidth * image.naturalHeight > MAX_PIXELS) throw Error('La imagen tiene demasiados píxeles.');
      return {image, ...dimensions(image.naturalWidth, image.naturalHeight, max), close: () => URL.revokeObjectURL(url)};
    } catch (error) { URL.revokeObjectURL(url); throw error; }
  }
  async function prepare(file, options) {
    if (!file || !file.size || file.size > MAX_INPUT) throw Error('Usa una imagen de hasta 20 MB.');
    if (file.type === 'image/svg+xml') throw Error('Usa una fotografía PNG, JPG o WebP.');
    const settings = {...presets.report, ...options}, meta = await header(file);
    if (meta && (!meta.width || !meta.height || meta.width * meta.height > MAX_PIXELS || Math.max(meta.width, meta.height) > 32000)) throw Error('La imagen tiene demasiados píxeles. Usa una copia de menor resolución.');
    if (meta && (!meta.orientation || meta.orientation === 1) && (!settings.forceJPEG || meta.type === 'image/jpeg') && Math.max(meta.width, meta.height) <= settings.maxDimension && file.size <= settings.maxBytes) {
      return {dataUrl: await toURL(file), size: file.size, ...meta, originalSize: file.size};
    }
    const decoded = await decode(file, meta, settings.maxDimension);
    try {
      let {width, height} = decoded, quality = settings.quality;
      const canvas = typeof OffscreenCanvas === 'function' ? new OffscreenCanvas(width, height) : document.createElement('canvas');
      const context = canvas.getContext('2d', {alpha: false});
      if (!context) throw Error('No se pudo preparar la imagen.');
      const encode = async () => {
        canvas.width = width; canvas.height = height; context.fillStyle = '#fff'; context.fillRect(0, 0, width, height); context.drawImage(decoded.image, 0, 0, width, height);
        if (canvas.convertToBlob) return canvas.convertToBlob({type: 'image/jpeg', quality});
        return new Promise((resolve, reject) => canvas.toBlob(blob => blob ? resolve(blob) : reject(Error('No se pudo comprimir la imagen.')), 'image/jpeg', quality));
      };
      let blob = await encode();
      for (let attempt = 0; blob.size > settings.maxBytes && attempt < 35; attempt++) {
        if (quality > settings.minQuality) quality = Math.max(settings.minQuality, quality - .06);
        else {
          if (Math.max(width, height) < 400) throw Error('Esta imagen no se puede reducir conservando una calidad útil. Elige otra imagen.');
          width = Math.max(1, Math.round(width * .88)); height = Math.max(1, Math.round(height * .88));
        }
        blob = await encode();
      }
      if (blob.size > settings.maxBytes) throw Error('No se pudo reducir el peso de esta imagen.');
      return {dataUrl: await toURL(blob), size: blob.size, type: 'image/jpeg', width, height, originalSize: file.size};
    } finally { decoded.close(); }
  }
  if (inWorker) {
    self.onmessage = async event => {
      const {id, file, options} = event.data;
      try { self.postMessage({id, result: await prepare(file, options)}); }
      catch (error) { self.postMessage({id, error: error.message || 'No se pudo preparar la imagen.'}); }
    };
    return;
  }
  let worker = null, sequence = 0; const jobs = new Map();
  function reset() {
    worker?.terminate(); worker = null;
    for (const job of jobs.values()) { clearTimeout(job.timer); job.reject(Error('El procesador de imágenes se reinició.')); }
    jobs.clear();
  }
  async function optimize(file, purpose = 'report', options = {}) {
    const settings = {...(presets[purpose] || presets.report), ...options};
    if (typeof Worker === 'function' && typeof OffscreenCanvas === 'function' && typeof createImageBitmap === 'function') {
      try {
        if (!worker) {
          worker = new Worker('music-media-612.js?v=6.1.2');
          worker.onerror = reset;
          worker.onmessage = event => {
            const {id, result, error} = event.data, job = jobs.get(id); if (!job) return;
            clearTimeout(job.timer); jobs.delete(id); error ? job.reject(Error(error)) : job.resolve(result);
          };
        }
        const id = ++sequence;
        return await new Promise((resolve, reject) => {
          const timer = setTimeout(reset, 60000); jobs.set(id, {resolve, reject, timer});
          try { worker.postMessage({id, file, options: settings}); } catch (error) { clearTimeout(timer); jobs.delete(id); reject(error); }
        });
      } catch { /* A browser without a usable worker still keeps the same byte limit. */ }
    }
    return prepare(file, settings);
  }
  async function pdfPhotoResource(source) {
    const comma = source.indexOf(','), mime = source.slice(5, source.indexOf(';'));
    const raw = Uint8Array.from(atob(source.slice(comma + 1)), c => c.charCodeAt(0));
    const photo = await optimize(new Blob([raw], {type: mime}), 'document', {forceJPEG: true});
    return {bytes: Uint8Array.from(atob(photo.dataUrl.split(',')[1]), c => c.charCodeAt(0)), alphaBytes: null, w: photo.width, h: photo.height, filter: 'DCTDecode'};
  }
  root.MusicMedia612 = {optimize, pdfPhotoResource};
})();
