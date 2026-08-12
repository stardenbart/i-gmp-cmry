const DEFAULT_KEY_BASE64 = "sCb2UdNCSu3RBEYLF6IG/18C6VAVuYftUhFB1lzRoyw=";

function base64ToUint8Array(base64: string): Uint8Array {
  const binaryString = atob(base64);
  const bytes = new Uint8Array(binaryString.length);
  for (let i = 0; i < binaryString.length; i++) {
    bytes[i] = binaryString.charCodeAt(i);
  }
  return bytes;
}

export function isEncryptedBase64(str: string | null | undefined): boolean {
  if (!str) return false;
  const trimmed = str.trim();
  return (
    trimmed.length >= 30 &&
    !trimmed.includes(" ") &&
    !trimmed.includes("/") &&
    !trimmed.match(/\.(jpg|jpeg|png|webp|gif|svg|bmp)$/i) &&
    (trimmed.endsWith("==") || trimmed.endsWith("=") || trimmed.includes("+"))
  );
}

export async function decryptAESGCM(cipherBase64: string, rawKeyBase64 = DEFAULT_KEY_BASE64): Promise<string> {
  if (!cipherBase64 || typeof window === "undefined" || !window.crypto?.subtle) {
    return cipherBase64;
  }
  try {
    const rawData = base64ToUint8Array(cipherBase64);
    if (rawData.length < 28) return cipherBase64; // 12 bytes IV + 16 bytes tag

    const iv = rawData.slice(0, 12);
    const ciphertextWithTag = rawData.slice(12);

    const keyBytes = base64ToUint8Array(rawKeyBase64);
    const keyBuffer = keyBytes.buffer.slice(keyBytes.byteOffset, keyBytes.byteOffset + keyBytes.byteLength) as ArrayBuffer;

    const cryptoKey = await window.crypto.subtle.importKey(
      "raw",
      keyBuffer,
      { name: "AES-GCM" },
      false,
      ["decrypt"]
    );

    const ivBuffer = iv.buffer.slice(iv.byteOffset, iv.byteOffset + iv.byteLength) as ArrayBuffer;
    const dataBuffer = ciphertextWithTag.buffer.slice(ciphertextWithTag.byteOffset, ciphertextWithTag.byteOffset + ciphertextWithTag.byteLength) as ArrayBuffer;

    const decryptedBuffer = await window.crypto.subtle.decrypt(
      { name: "AES-GCM", iv: ivBuffer },
      cryptoKey,
      dataBuffer
    );

    return new TextDecoder().decode(decryptedBuffer);
  } catch (err) {
    return cipherBase64;
  }
}

/**
 * Recursively inspects API response data and decrypts any encrypted base64 strings
 * found in fields like keterangan, image_url, file_name, etc.
 */
export async function decryptApiResponseData<T>(data: T): Promise<T> {
  if (!data || typeof data !== "object") return data;

  if (Array.isArray(data)) {
    return (await Promise.all(data.map((item) => decryptApiResponseData(item)))) as unknown as T;
  }

  const obj = { ...data } as any;
  const targetKeys = ["keterangan", "image_url", "file_name", "ImageUrl", "FileName", "Keterangan"];

  for (const key of Object.keys(obj)) {
    const val = obj[key];
    if (typeof val === "string" && targetKeys.includes(key) && isEncryptedBase64(val)) {
      obj[key] = await decryptAESGCM(val);
    } else if (val && typeof val === "object") {
      obj[key] = await decryptApiResponseData(val);
    }
  }

  return obj as T;
}
