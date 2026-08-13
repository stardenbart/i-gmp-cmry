/**
 * ╔══════════════════════════════════════════════════════════════════════╗
 * ║         SYSTEM-AUDIT — k6 LOAD TEST SCRIPT                         ║
 * ║   Simulasi: Login → Dashboard → Inspeksi → Draft Sync → HEI        ║
 * ╚══════════════════════════════════════════════════════════════════════╝
 *
 * CARA PAKAI:
 *   1. Jalankan backend dan pastikan bisa diakses di BASE_URL di bawah
 *   2. Jalankan:  k6 run load_test.js
 *   3. Atau dengan override username/password:
 *      k6 run -e USERNAME=admin_sentul -e PASSWORD=your_pass load_test.js
 */

import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// ── Konfigurasi ─────────────────────────────────────────────────────────────
const BASE_URL    = __ENV.BASE_URL    || 'http://localhost:8080/api/v1';
const USERNAME    = __ENV.USERNAME    || 'admin_sentul';    // ← Ganti sesuai user Anda
const PASSWORD    = __ENV.PASSWORD    || 'admin123';     // ← Ganti sesuai password Anda
const PLANT_CODE  = __ENV.PLANT_CODE  || 'PLT-SENTUL';
const KAWASAN_ID  = __ENV.KAWASAN_ID  || 'K001';
const ASPEK_IDS   = ['ASP001', 'ASP002', 'ASP003'];

// ── Custom Metrics ───────────────────────────────────────────────────────────
const errorRate       = new Rate('custom_error_rate');
const draftSyncTrend  = new Trend('draft_sync_ms',    true);
const checklistTrend  = new Trend('checklist_load_ms', true);
const loginErrors     = new Counter('login_errors');

// ── Skenario Beban ───────────────────────────────────────────────────────────
export const options = {
  stages: [
    { duration: '20s', target: 5   },  // Warmup     : 5 user
    { duration: '30s', target: 20  },  // Normal     : 20 user (1 plant kecil)
    { duration: '1m',  target: 50  },  // Peak       : 50 user (multi-plant)
    { duration: '30s', target: 100 },  // Stress     : 100 user concurrent
    { duration: '30s', target: 200 },  // Break-point: 200 user (temukan batas)
    { duration: '20s', target: 0   },  // Ramp-down
  ],
  thresholds: {
    // Lulus jika:
    http_req_failed:      ['rate<0.05'],    // Error rate < 5%
    http_req_duration:    ['p(95)<1000'],   // 95% request < 1 detik
    custom_error_rate:    ['rate<0.03'],    // Custom error < 3%
    draft_sync_ms:        ['p(95)<800'],    // Draft sync < 800ms
    checklist_load_ms:    ['p(95)<1500'],   // Checklist (berat) < 1.5 detik
  },
};

// ── Setup: Login Sekali untuk Dapat Token ───────────────────────────────────
export function setup() {
  console.log(`🚀 Setup: Login sebagai ${USERNAME} ke ${BASE_URL}`);

  const res = http.post(`${BASE_URL}/auth/login`, JSON.stringify({
    username: USERNAME,
    password: PASSWORD,
  }), {
    headers: { 'Content-Type': 'application/json' },
  });

  if (res.status !== 200) {
    console.error(`❌ Login gagal! Status: ${res.status} Body: ${res.body}`);
    loginErrors.add(1);
    return { token: null, inspectionIds: [] };
  }

  const body = JSON.parse(res.body);
  const token = body.data?.token || body.token || '';
  console.log(`✅ Login berhasil. Token: ${token.substring(0, 30)}...`);

  // Ambil daftar inspection ID yang tersedia untuk simulasi read
  const inspRes = http.get(`${BASE_URL}/inspections?limit=10`, {
    headers: { 'Authorization': `Bearer ${token}` },
  });
  let inspectionIds = ['INSP-260812-f99cce7e']; // fallback ID
  if (inspRes.status === 200) {
    const inspBody = JSON.parse(inspRes.body);
    const ids = (inspBody.data?.items || inspBody.items || []).map(i => i.inspection_id || i.InspectionID);
    if (ids.length > 0) inspectionIds = ids;
  }
  console.log(`📋 Ditemukan ${inspectionIds.length} inspeksi untuk simulasi`);

  return { token, inspectionIds };
}

// ── Main Virtual User Logic ──────────────────────────────────────────────────
export default function (data) {
  const { token, inspectionIds } = data;

  // Jika setup gagal (tidak ada token), skip test
  if (!token) {
    console.warn('⚠️  Tidak ada token — pastikan backend jalan dan credentials benar');
    sleep(5);
    return;
  }

  const headers = {
    'Content-Type':  'application/json',
    'Authorization': `Bearer ${token}`,
  };

  // Pilih inspection ID secara random (simulasi user buka inspeksi berbeda-beda)
  const inspId = inspectionIds[Math.floor(Math.random() * inspectionIds.length)];
  const aspId  = ASPEK_IDS[Math.floor(Math.random() * ASPEK_IDS.length)];

  // ────────────────────────────────────────────────────────────────────────
  // GRUP 1: Dashboard & Master Data (paling sering dipanggil)
  // ────────────────────────────────────────────────────────────────────────
  group('1_dashboard', () => {
    // Status kawasan (di-cache backend, harusnya < 50ms setelah hit pertama)
    let r = http.get(`${BASE_URL}/inspeksi/${KAWASAN_ID}/status`, { headers });
    check(r, { '✓ status kawasan': (r) => r.status === 200 });
    errorRate.add(r.status !== 200);
  });

  sleep(randomPause(0.5, 1.5));

  // ────────────────────────────────────────────────────────────────────────
  // GRUP 2: Master Data HEI (heavy saat cold, di-cache 1 jam)
  // ────────────────────────────────────────────────────────────────────────
  group('2_master_data', () => {
    // Categories (di-cache 1 jam, harusnya cepat setelah warmup)
    let r1 = http.get(`${BASE_URL}/master/hei/categories`, { headers });
    check(r1, { '✓ hei categories': (r) => r.status === 200 });
    errorRate.add(r1.status !== 200);

    sleep(0.3);

    // HEI list (di-cache 1 jam)
    let r2 = http.get(`${BASE_URL}/master/hei?page=1&limit=100`, { headers });
    check(r2, { '✓ hei list': (r) => r.status === 200 });
    errorRate.add(r2.status !== 200);
  });

  sleep(randomPause(0.5, 1));

  // ────────────────────────────────────────────────────────────────────────
  // GRUP 3: Buka Detail Inspeksi (JOIN 4 tabel, di-cache 30 detik)
  // ────────────────────────────────────────────────────────────────────────
  group('3_inspection_detail', () => {
    let r = http.get(`${BASE_URL}/inspections/${inspId}`, { headers });
    check(r, { '✓ inspection detail': (r) => r.status === 200 });
    errorRate.add(r.status !== 200 && r.status !== 404);
  });

  sleep(randomPause(1, 2));

  // ────────────────────────────────────────────────────────────────────────
  // GRUP 4: Load Checklist (paling berat — ini sering jadi bottleneck)
  // ────────────────────────────────────────────────────────────────────────
  group('4_checklist', () => {
    const start = Date.now();
    let r = http.get(`${BASE_URL}/inspections/${inspId}/checklist`, { headers });
    checklistTrend.add(Date.now() - start);
    check(r, { '✓ checklist load': (r) => r.status === 200 || r.status === 404 });
    errorRate.add(r.status !== 200 && r.status !== 404);
  });

  sleep(randomPause(1, 2));

  // ────────────────────────────────────────────────────────────────────────
  // GRUP 5: Get Redis Draft (batch read dari Redis)
  // ────────────────────────────────────────────────────────────────────────
  group('5_get_drafts', () => {
    let r = http.get(`${BASE_URL}/inspeksi/${inspId}/drafts`, { headers });
    check(r, { '✓ get drafts': (r) => r.status === 200 || r.status === 404 });
    errorRate.add(r.status !== 200 && r.status !== 404);
  });

  sleep(randomPause(0.5, 1.5));

  // ────────────────────────────────────────────────────────────────────────
  // GRUP 6: Draft Sync (write burst — paling kritis untuk Redis throughput)
  //         Mensimulasikan auditor mengetik nilai + keterangan per aspek
  // ────────────────────────────────────────────────────────────────────────
  group('6_draft_sync_write', () => {
    const payload = JSON.stringify({
      data: {
        [`DTL-001_URN-00${randomInt(1, 5)}`]: {
          checking:   Math.random() > 0.3 ? 'OK' : 'NG',
          nilai:      Math.random() > 0.3 ? 100 : 0,
          keterangan: `Load test note ${Date.now()}`,
          photos:     [],
        },
        [`DTL-002_URN-00${randomInt(1, 3)}`]: {
          checking:   'OK',
          nilai:      100,
          keterangan: '',
          photos:     [],
        },
      },
      skor: randomInt(60, 100),
    });

    const start = Date.now();
    let r = http.put(
      `${BASE_URL}/inspeksi/${KAWASAN_ID}/${aspId}`,
      payload,
      { headers: { ...headers, 'X-Lock-Token': 'load-test-no-lock' } }
    );
    draftSyncTrend.add(Date.now() - start);

    // 200=OK, 403=tidak punya lock (expected), 404=aspek tidak ada (expected)
    const ok = [200, 403, 404].includes(r.status);
    check(r, { '✓ draft sync accepted': () => ok });
    errorRate.add(!ok);
  });

  sleep(randomPause(1, 3)); // User membaca/mengisi form sebelum aksi berikutnya
}

// ── Teardown: Laporan Ringkas ────────────────────────────────────────────────
export function teardown(data) {
  console.log('\n══════════════════════════════════════════════');
  console.log('  Load Test Selesai!');
  console.log('  Cek output di atas untuk thresholds pass/fail');
  console.log('  Fokus pada: p(95) latency & error rate');
  console.log('══════════════════════════════════════════════\n');
}

// ── Helper ───────────────────────────────────────────────────────────────────
function randomPause(min, max) {
  return Math.random() * (max - min) + min;
}

function randomInt(min, max) {
  return Math.floor(Math.random() * (max - min + 1)) + min;
}
