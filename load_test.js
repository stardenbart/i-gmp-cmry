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
 *      k6 run -e USERNAME=admin_sentul -e PASSWORD=admin123 load_test.js
 */

import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// ── Konfigurasi ─────────────────────────────────────────────────────────────
const BASE_URL    = __ENV.BASE_URL    || 'http://localhost:8080/api/v1';
const USERNAME    = __ENV.USERNAME    || 'admin_sentul';
const PASSWORD    = __ENV.PASSWORD    || 'admin123';
const PLANT_CODE  = __ENV.PLANT_CODE  || 'PLT-SENTUL';
// Real IDs sesuai master data yang sudah di-seed
const KAWASAN_ID  = __ENV.KAWASAN_ID  || 'KWS-001';
const ASPEK_IDS   = ['ASP-001', 'ASP-002', 'ASP-003', 'ASP-004', 'ASP-005'];

// ── Custom Metrics ───────────────────────────────────────────────────────────
const errorRate          = new Rate('custom_error_rate');
const draftSyncTrend     = new Trend('draft_sync_ms',     true);
const checklistTrend     = new Trend('checklist_load_ms', true);
const loginErrors        = new Counter('login_errors');

// Per-status counters — untuk membedah gap http_req_failed vs checks_failed
const draftSyncOk        = new Counter('draft_sync_200');   // write sukses, data tersimpan
const draftSyncConflict  = new Counter('draft_sync_409');   // lock conflict (user lain)
const draftSyncForbidden = new Counter('draft_sync_403');   // forbidden
const draftSyncOther     = new Counter('draft_sync_other'); // 5xx / timeout
const getDraftsOk        = new Counter('get_drafts_200');   // cache hit
const getDraftsEmpty     = new Counter('get_drafts_404');   // kosong (belum pernah write)
const getDraftsTimeout   = new Counter('get_drafts_503');   // fail-fast timeout
const getDraftsOther     = new Counter('get_drafts_other'); // anomali lain

// ── Skenario Beban ───────────────────────────────────────────────────────────
export const options = {
  stages: [
    { duration: '20s', target: 5   },  // Warmup     : 5 user
    { duration: '30s', target: 20  },  // Normal     : 20 user
    { duration: '1m',  target: 50  },  // Peak       : 50 user
    { duration: '30s', target: 100 },  // Stress     : 100 user concurrent
    { duration: '30s', target: 200 },  // Break-point: 200 user
    { duration: '20s', target: 0   },  // Ramp-down
  ],
  thresholds: {
    http_req_failed:      ['rate<0.05'],    // Error rate < 5%
    http_req_duration:    ['p(95)<1000'],   // 95% request < 1 detik
    custom_error_rate:    ['rate<0.03'],    // Custom error < 3%
    draft_sync_ms:        ['p(95)<800'],    // Draft sync < 800ms
    checklist_load_ms:    ['p(95)<1500'],   // Checklist < 1.5 detik
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

  // Ambil daftar inspection ID yang tersedia (gunakan limit besar agar dapat banyak variasi)
  const inspRes = http.get(`${BASE_URL}/inspections?limit=50`, {
    headers: { 'Authorization': `Bearer ${token}` },
  });
  // Fallback ke ID seed yang baru (valid setelah reseed — 63 inspection headers)
  let inspectionIds = [
    'INSP-LT-001-01', 'INSP-LT-001-02', 'INSP-LT-001-03',
    'INSP-LT-002-01', 'INSP-LT-002-02', 'INSP-LT-002-03',
    'INSP-LT-003-01', 'INSP-LT-003-02', 'INSP-LT-003-03',
    'INSP-LT-004-01', 'INSP-LT-004-02', 'INSP-LT-005-01',
    'INSP-LT-006-01', 'INSP-LT-007-01', 'INSP-LT-008-01',
    'INSP-LT-009-01', 'INSP-LT-010-01', 'INSP-LT-011-01',
    'INSP-LT-012-01', 'INSP-LT-013-01', 'INSP-LT-014-01',
  ];
  if (inspRes.status === 200) {
    const inspBody = JSON.parse(inspRes.body);
    const ids = (inspBody.data?.items || inspBody.items || []).map(i => i.inspection_id || i.InspectionID || i.id).filter(Boolean);
    if (ids.length > 0) {
      inspectionIds = ids;
      console.log(`📋 Fetched ${ids.length} inspection IDs from API`);
    } else {
      console.log(`⚠️  API returned 0 inspections, using fallback IDs (${inspectionIds.length} total)`);
    }
  } else {
    console.warn(`⚠️  GET /inspections failed [${inspRes.status}], using fallback IDs`);
  }
  console.log(`📋 Using ${inspectionIds.length} inspection IDs: ${inspectionIds.slice(0,3).join(', ')}...`);

  return { token, inspectionIds };
}

// ── Main Virtual User Logic ──────────────────────────────────────────────────
export default function (data) {
  const { token, inspectionIds } = data;

  if (!token) {
    console.warn('⚠️  Tidak ada token — pastikan backend jalan dan credentials benar');
    sleep(5);
    return;
  }

  const headers = {
    'Content-Type':  'application/json',
    'Authorization': `Bearer ${token}`,
  };

  const inspId = inspectionIds[Math.floor(Math.random() * inspectionIds.length)];
  const aspId  = ASPEK_IDS[Math.floor(Math.random() * ASPEK_IDS.length)];

  // 1. Dashboard
  group('1_dashboard', () => {
    let r = http.get(`${BASE_URL}/inspeksi/${KAWASAN_ID}/status`, { headers });
    const isOk = check(r, { '✓ status kawasan': (r) => r.status === 200 });
    if (!isOk) console.warn(`Dashboard status failed: ${r.status} ${r.body}`);
    errorRate.add(!isOk);
  });

  sleep(randomPause(0.5, 1.5));

  // 2. Master Data
  group('2_master_data', () => {
    let r1 = http.get(`${BASE_URL}/master/hei/categories`, { headers });
    check(r1, { '✓ hei categories': (r) => r.status === 200 });
    errorRate.add(r1.status !== 200);

    sleep(0.3);

    let r2 = http.get(`${BASE_URL}/master/hei?page=1&limit=100`, { headers });
    check(r2, { '✓ hei list': (r) => r.status === 200 });
    errorRate.add(r2.status !== 200);
  });

  sleep(randomPause(0.5, 1));

  // 3. Inspection Detail
  group('3_inspection_detail', () => {
    let r = http.get(`${BASE_URL}/inspections/${inspId}`, { headers });
    const isOk = check(r, { '✓ inspection detail': (r) => r.status === 200 });
    if (!isOk) console.warn(`Inspection detail failed [${r.status}]: ${r.body}`);
    errorRate.add(!isOk);
  });

  sleep(randomPause(1, 2));

  // 4. Load Checklist (Hindari Survivorship Bias: Hanya catat Trend jika OK)
  group('4_checklist', () => {
    const start = Date.now();
    let r = http.get(`${BASE_URL}/inspections/${inspId}/checklist`, { headers });
    const duration = Date.now() - start;
    
    const isOk = check(r, { '✓ checklist load': (r) => r.status === 200 });
    if (isOk) {
      checklistTrend.add(duration);
    } else {
      console.warn(`Checklist load failed [${r.status}]: ${r.body}`);
    }
    errorRate.add(!isOk);
  });

  sleep(randomPause(1, 2));

  // 5. Get Redis Draft
  // Scope KAWASAN_ID (match dengan write scope di group 6)
  // Key write: state:aspek:KWS-001:ASP-xxx  →  pattern read: state:aspek:KWS-001:*  ✓
  // 503 TIDAK ditoleransi: dengan scope benar, cache hit seharusnya 200.
  // Kalau masih 503 = ada masalah latensi residual yang perlu terlihat di checks_failed.
  // responseCallback: memberitahu k6 bahwa 404 bukan failure (Redis kosong = valid state)
  group('5_get_drafts', () => {
    let r = http.get(`${BASE_URL}/inspeksi/${KAWASAN_ID}/drafts`, {
      headers,
      responseCallback: http.expectedStatuses(200, 404), // 404 ok, bukan kegagalan
    });
    const isOk = check(r, { '✓ get drafts': (r) => r.status === 200 || r.status === 404 });
    errorRate.add(!isOk);

    // Per-status breakdown
    if      (r.status === 200) getDraftsOk.add(1);
    else if (r.status === 404) getDraftsEmpty.add(1);
    else if (r.status === 503) getDraftsTimeout.add(1);
    else                       getDraftsOther.add(1);
  });

  sleep(randomPause(0.5, 1.5));

  // 6. Draft Sync — alur realistis: Acquire Lock → SaveAspek
  //
  // Latar belakang ValidateLock:
  //   - Lock bebas / expired → auto-grant ke caller (return 200)
  //   - Lock dipegang SAME userID → auto-refresh TTL (return 200)
  //   - Lock dipegang user LAIN → 409 Conflict
  //   Karena semua VU login sebagai admin_sentul yang SAMA,
  //   ValidateLock praktis selalu grant/refresh → draft write benar-benar berjalan.
  //   Token 'load-test-no-lock' valid karena lock bebas akan auto-granted.
  //
  // responseCallback pada PUT memberitahu k6 bahwa 409 bukan kegagalan yang perlu
  // dihitung di http_req_failed — ini adalah respons expected saat lock conflict.
  group('6_draft_sync_write', () => {
    const payload = JSON.stringify({
      data: {
        [`DTL-001_URN-00${randomInt(1, 5)}`]: {
          checking:   Math.random() > 0.3 ? 'OK' : 'NG',
          nilai:      Math.random() > 0.3 ? 100 : 0,
          keterangan: `Load test note ${Date.now()}`,
          photos:     [],
        },
      },
      skor: randomInt(60, 100),
    });

    const start = Date.now();
    let r = http.put(
      `${BASE_URL}/inspeksi/${KAWASAN_ID}/${aspId}`,
      payload,
      {
        headers: { ...headers, 'X-Lock-Token': 'load-test-no-lock' },
        // Beritahu k6: 409 adalah respons expected (lock conflict), bukan kegagalan
        responseCallback: http.expectedStatuses(200, 409),
      }
    );
    const elapsed = Date.now() - start;

    // Hanya catat trend untuk write yang benar-benar berhasil (200)
    if (r.status === 200) draftSyncTrend.add(elapsed);

    // draft_sync dianggap "ok" selama tidak ada error server (5xx) atau timeout
    // 409 = lock conflict = respons valid dari sistem yang bekerja benar
    const ok = r.status === 200 || r.status === 409;
    check(r, { '✓ draft sync accepted': () => ok });
    errorRate.add(!ok);

    // Per-status breakdown
    if      (r.status === 200) draftSyncOk.add(1);
    else if (r.status === 409) draftSyncConflict.add(1);
    else if (r.status === 403) draftSyncForbidden.add(1);
    else                       draftSyncOther.add(1);

    if (!ok) console.warn(`Draft sync failed [${r.status}]: ${r.body.substring(0, 120)}`);
  });

  sleep(randomPause(1, 3));
}

export function teardown(data) {
  console.log('\n══════════════════════════════════════════════');
  console.log('  Load Test Selesai!');
  console.log('══════════════════════════════════════════════\n');
}

function randomPause(min, max) {
  return Math.random() * (max - min) + min;
}

function randomInt(min, max) {
  return Math.floor(Math.random() * (max - min + 1)) + min;
}
