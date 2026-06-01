import http from 'k6/http';
import { check, sleep } from 'k6';

// Configuráveis via ENV:
// BASE_URL - URL base da aplicação (default http://localhost:8000)
// ACCOUNTS - número de contas / VUs a simular (default 10)
// EMPLOYEES_PER_ACCOUNT - quantos funcionários cada conta cria (default 5)
// SLEEP_MS - tempo entre requests em ms (default 200)

// Calcular `vus` e `iterations` garantindo `iterations >= vus` para evitar erro do k6
const VUS = __ENV.ACCOUNTS ? parseInt(__ENV.ACCOUNTS) : 10;
let ITER = __ENV.ITERATIONS ? parseInt(__ENV.ITERATIONS) : VUS;
if (isNaN(ITER) || ITER < VUS) {
	ITER = VUS;
}

export let options = {
	vus: VUS,
	iterations: ITER,
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8000';
const EMPLOYEES_PER_ACCOUNT = __ENV.EMPLOYEES_PER_ACCOUNT ? parseInt(__ENV.EMPLOYEES_PER_ACCOUNT) : 5;
const SLEEP_MS = __ENV.SLEEP_MS ? parseInt(__ENV.SLEEP_MS) : 200;

function formEncode(obj) {
	return Object.keys(obj)
		.map(k => `${encodeURIComponent(k)}=${encodeURIComponent(obj[k])}`)
		.join('&');
}

export default function () {
	// Cada VU representa uma conta distinta


	// helpers para nomes e tokens
	function randInt(max) { return Math.floor(Math.random() * max); }
	function randomHex(len) {
		const chars = 'abcdef0123456789';
		let s = '';
		for (let i = 0; i < len; i++) s += chars[Math.floor(Math.random() * chars.length)];
		return s;
	}
	function choice(arr) { return arr[randInt(arr.length)]; }

	const firstNames = ['Ana', 'Bruno', 'Carla', 'Diego', 'Eduarda', 'Felipe', 'Gabriela', 'Heitor', 'Isabela', 'João', 'Larissa', 'Marcos', 'Natália', 'Otávio', 'Pietra', 'Rafael', 'Sofia', 'Thiago', 'Victor', 'Yara'];
	const lastNames = ['Silva', 'Souza', 'Oliveira', 'Pereira', 'Costa', 'Gomes', 'Ribeiro', 'Almeida', 'Fernandes', 'Carvalho', 'Rocha', 'Lima', 'Martins', 'Araújo', 'Mendes'];
	const techPositions = ['Software Engineer', 'Frontend Engineer', 'Backend Engineer', 'Fullstack Engineer', 'DevOps Engineer', 'SRE', 'QA Engineer', 'Data Scientist', 'Product Manager', 'UX Designer', 'Mobile Engineer', 'Security Engineer'];

	const vu = __VU;
	const unique = `${Date.now()}-${vu}-${Math.floor(Math.random() * 10000)}`;
	// users com nomes "não normais": username/token hex e email baseado no token
	const username = `u_${randomHex(12)}_${vu}`;
	const password = `P!${randomHex(10)}`;
	const email = `${username}@example.com`;

	// Registrar conta
	let registerRes = http.post(
		`${BASE_URL}/register`,
		formEncode({ username, password, email }),
		{ headers: { 'Content-Type': 'application/x-www-form-urlencoded' } }
	);

	check(registerRes, {
		'registered (status 200 or 302)': r => r.status === 200 || r.status === 302,
	});

	sleep(SLEEP_MS / 1000);

	// Fazer login para obter cookie de sessão
	let loginRes = http.post(
		`${BASE_URL}/login`,
		formEncode({ username, email, password }),
		{ headers: { 'Content-Type': 'application/x-www-form-urlencoded' } }
	);

	check(loginRes, {
		'logged in (redirect 302)': r => r.status === 302 || r.status === 200,
	});

	sleep(SLEEP_MS / 1000);

	// Criar funcionários
	for (let i = 0; i < EMPLOYEES_PER_ACCOUNT; i++) {
		const full_name = `${choice(firstNames)} ${choice(lastNames)}${i % 10 === 0 ? ' Jr.' : ''}`;
		const emp_email = `emp_${randomHex(6)}_${i}_${vu}@example.com`;
		const position = choice(techPositions);

		let createRes = http.post(
			`${BASE_URL}/employees`,
			formEncode({ full_name, email: emp_email, position }),
			{ headers: { 'Content-Type': 'application/x-www-form-urlencoded' } }
		);

		// O handler redireciona com 302 para /employees em caso de sucesso
		check(createRes, {
			'employee created (302)': r => r.status === 302 || r.status === 200,
		});

		sleep(SLEEP_MS / 1000);
	}

	// Opcional: logout
	http.get(`${BASE_URL}/logout`);
	sleep(SLEEP_MS / 1000);
}

