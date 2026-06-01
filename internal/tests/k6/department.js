import http from 'k6/http';
import { check, sleep } from 'k6';

// Script k6 para criar departamentos.
// Variáveis ENV:
// BASE_URL, ACCOUNTS, EMPLOYEES_PER_ACCOUNT (não usado aqui), SLEEP_MS, ITERATIONS

const VUS = __ENV.ACCOUNTS ? parseInt(__ENV.ACCOUNTS) : 5;
let ITER = __ENV.ITERATIONS ? parseInt(__ENV.ITERATIONS) : VUS;
if (isNaN(ITER) || ITER < VUS) ITER = VUS;

export let options = {
  vus: VUS,
  iterations: ITER,
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8000';
const SLEEP_MS = __ENV.SLEEP_MS ? parseInt(__ENV.SLEEP_MS) : 200;

function formEncode(obj) {
  return Object.keys(obj).map(k => `${encodeURIComponent(k)}=${encodeURIComponent(obj[k])}`).join('&');
}

function randInt(max) { return Math.floor(Math.random() * max); }
function randomHex(len) { const chars = 'abcdef0123456789'; let s = ''; for (let i=0;i<len;i++) s += chars[Math.floor(Math.random()*chars.length)]; return s; }
function choice(arr) { return arr[randInt(arr.length)]; }

const deptNames = [
  'Platform', 'Infrastructure', 'Cloud Services', 'Data Engineering', 'AI Research', 'Mobile', 'Web', 'Security', 'Developer Experience', 'Quality Engineering', 'Product', 'Design'
];

export default function() {
  // gerar user "não normal"
  const vu = __VU;
  const token = randomHex(12) + '-' + vu;
  const username = `u_${token}`;
  const password = `P!${randomHex(10)}`;
  const email = `${username}@example.com`;

  // register
  let res = http.post(`${BASE_URL}/register`, formEncode({ username, password, email }), { headers: { 'Content-Type': 'application/x-www-form-urlencoded' } });
  check(res, { 'registered': r => r.status === 200 || r.status === 302 });
  sleep(SLEEP_MS/1000);

  // login
  res = http.post(`${BASE_URL}/login`, formEncode({ username, email, password }), { headers: { 'Content-Type': 'application/x-www-form-urlencoded' } });
  check(res, { 'logged in': r => r.status === 200 || r.status === 302 });
  sleep(SLEEP_MS/1000);

  // criar departamento (sem boss_id para simplicidade)
  const name = `${choice(deptNames)} Dept ${randomHex(4)}`;
  const code = `DPT-${randomHex(3)}-${vu}`;

  res = http.post(`${BASE_URL}/department`, formEncode({ code, name }), { headers: { 'Content-Type': 'application/x-www-form-urlencoded' } });
  check(res, { 'dept created': r => r.status === 200 || r.status === 302 });

  sleep(SLEEP_MS/1000);

  // opcional logout
  http.get(`${BASE_URL}/logout`);
  sleep(SLEEP_MS/1000);
}
