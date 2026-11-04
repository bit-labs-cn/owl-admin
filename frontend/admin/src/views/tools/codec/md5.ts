const rounds = [
  [7, 12, 17, 22],
  [5, 9, 14, 20],
  [4, 11, 16, 23],
  [6, 10, 15, 21]
] as const;

const sine = new Uint32Array(64);
for (let i = 0; i < 64; i += 1) {
  sine[i] = Math.floor(Math.abs(Math.sin(i + 1)) * 0x100000000);
}

function rotateLeft(value: number, bits: number) {
  return ((value << bits) | (value >>> (32 - bits))) >>> 0;
}

function word(bytes: Uint8Array, offset: number) {
  return (
    (bytes[offset] | (bytes[offset + 1] << 8) | (bytes[offset + 2] << 16) | (bytes[offset + 3] << 24)) >>> 0
  );
}

function hexWord(value: number) {
  const bytes = [value & 0xff, (value >>> 8) & 0xff, (value >>> 16) & 0xff, (value >>> 24) & 0xff];
  return bytes.map(item => item.toString(16).padStart(2, "0")).join("");
}

/** md5 计算 UTF-8 文本的 MD5，结果为小写十六进制。 */
export function md5(message: string) {
  const data = new TextEncoder().encode(message);
  const bitLength = data.length * 8;
  const paddedLength = (((data.length + 8) >> 6) + 1) << 6;
  const bytes = new Uint8Array(paddedLength);
  bytes.set(data);
  bytes[data.length] = 0x80;
  const lengthView = new DataView(bytes.buffer);
  lengthView.setUint32(paddedLength - 8, bitLength >>> 0, true);
  lengthView.setUint32(paddedLength - 4, Math.floor(bitLength / 0x100000000), true);

  let a0 = 0x67452301;
  let b0 = 0xefcdab89;
  let c0 = 0x98badcfe;
  let d0 = 0x10325476;

  for (let offset = 0; offset < bytes.length; offset += 64) {
    let a = a0;
    let b = b0;
    let c = c0;
    let d = d0;
    for (let i = 0; i < 64; i += 1) {
      let mixed = 0;
      let index = 0;
      if (i < 16) {
        mixed = (b & c) | (~b & d);
        index = i;
      } else if (i < 32) {
        mixed = (d & b) | (~d & c);
        index = (5 * i + 1) % 16;
      } else if (i < 48) {
        mixed = b ^ c ^ d;
        index = (3 * i + 5) % 16;
      } else {
        mixed = c ^ (b | ~d);
        index = (7 * i) % 16;
      }
      const sum = (a + (mixed >>> 0) + sine[i] + word(bytes, offset + index * 4)) >>> 0;
      const next = (b + rotateLeft(sum, rounds[i >> 4][i & 3])) >>> 0;
      a = d;
      d = c;
      c = b;
      b = next;
    }
    a0 = (a0 + a) >>> 0;
    b0 = (b0 + b) >>> 0;
    c0 = (c0 + c) >>> 0;
    d0 = (d0 + d) >>> 0;
  }

  return hexWord(a0) + hexWord(b0) + hexWord(c0) + hexWord(d0);
}
