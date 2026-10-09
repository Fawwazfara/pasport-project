document.addEventListener('DOMContentLoaded', function () {

	/* ---------- utilitas bersama ---------- */

	function setError(id, msg, ok) {
		var el = document.getElementById(id);
		if (!el) return;
		el.textContent = msg || '';
		el.style.display = msg ? 'block' : 'none';
		el.classList.toggle('ok', !!ok);
	}

	function clearErrors(ids) {
		ids.forEach(function (id) {
			setError(id, '');
		});
	}

	function setLoading(btn, text) {
		if (text) {
			btn.disabled = true;
			btn.textContent = text;
		} else {
			btn.disabled = false;
			btn.textContent = btn.dataset.label;
		}
	}

	function formatRupiah(n) {
		return 'Rp ' + n.toLocaleString('id-ID');
	}

	function apiFetch(url, options) {
		return fetch(url, options).then(function (res) {
			return res
				.json()
				.catch(function () {
					return {};
				})
				.then(function (data) {
					if (!res.ok) throw new Error(data.error || 'Terjadi kesalahan');
					return data;
				});
		});
	}

	/* ---------- auth ---------- */

	var authMain = document.getElementById('authMain');
	var appMain = document.getElementById('appMain');
	var userArea = document.getElementById('userArea');
	var userName = document.getElementById('userName');
	var authViews = ['loginView', 'registerView', 'resetView'].map(function (id) {
		return document.getElementById(id);
	});

	function showAuth(viewId) {
		authMain.hidden = false;
		appMain.hidden = true;
		userArea.hidden = true;
		authViews.forEach(function (v) {
			v.classList.toggle('active', v.id === viewId);
		});
		window.scrollTo(0, 0);
	}

	function showApp(user) {
		authMain.hidden = true;
		appMain.hidden = false;
		userArea.hidden = false;
		userName.textContent = user.nama || user.username;
		window.scrollTo(0, 0);
	}

	document.querySelectorAll('[data-goto]').forEach(function (btn) {
		btn.addEventListener('click', function () {
			showAuth(btn.dataset.goto);
		});
	});

	var loginBtn = document.getElementById('loginBtn');
	loginBtn.dataset.label = loginBtn.textContent;
	document.getElementById('loginForm').addEventListener('submit', function (e) {
		e.preventDefault();
		clearErrors(['loginUsernameError', 'loginPasswordError', 'loginError']);

		var username = document.getElementById('loginUsername').value.trim();
		var password = document.getElementById('loginPassword').value;

		var ok = true;
		if (!username) {
			setError('loginUsernameError', 'Username wajib diisi');
			ok = false;
		}
		if (!password) {
			setError('loginPasswordError', 'Kata sandi wajib diisi');
			ok = false;
		}
		if (!ok) return;

		setLoading(loginBtn, 'Memproses...');
		apiFetch('/api/auth/login', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ username: username, password: password })
		})
			.then(function (user) {
				document.getElementById('loginForm').reset();
				showApp(user);
			})
			.catch(function (err) {
				setError('loginError', err.message);
			})
			.finally(function () {
				setLoading(loginBtn, null);
			});
	});

	var registerBtn = document.getElementById('registerBtn');
	registerBtn.dataset.label = registerBtn.textContent;
	document.getElementById('registerForm').addEventListener('submit', function (e) {
		e.preventDefault();
		clearErrors([
			'registerNamaError',
			'registerUsernameError',
			'registerPasswordError',
			'registerConfirmError',
			'registerError'
		]);

		var nama = document.getElementById('registerNama').value.trim();
		var username = document.getElementById('registerUsername').value.trim();
		var password = document.getElementById('registerPassword').value;
		var confirm = document.getElementById('registerConfirm').value;

		var ok = true;
		if (!nama) {
			setError('registerNamaError', 'Nama lengkap wajib diisi');
			ok = false;
		}
		if (!/^[a-zA-Z0-9._-]{3,32}$/.test(username)) {
			setError('registerUsernameError', 'Username 3-32 karakter (huruf, angka, . _ -)');
			ok = false;
		}
		if (password.length < 8) {
			setError('registerPasswordError', 'Kata sandi minimal 8 karakter');
			ok = false;
		}
		if (confirm !== password) {
			setError('registerConfirmError', 'Ulangi kata sandi dengan benar');
			ok = false;
		}
		if (!ok) return;

		setLoading(registerBtn, 'Memproses...');
		apiFetch('/api/auth/register', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ username: username, nama: nama, password: password })
		})
			.then(function (user) {
				document.getElementById('registerForm').reset();
				showApp(user);
			})
			.catch(function (err) {
				setError('registerError', err.message);
			})
			.finally(function () {
				setLoading(registerBtn, null);
			});
	});

	var resetBtn = document.getElementById('resetBtn');
	resetBtn.dataset.label = resetBtn.textContent;
	document.getElementById('resetForm').addEventListener('submit', function (e) {
		e.preventDefault();
		clearErrors(['resetUsernameError', 'resetPasswordError', 'resetConfirmError', 'resetError']);

		var username = document.getElementById('resetUsername').value.trim();
		var password = document.getElementById('resetPassword').value;
		var confirm = document.getElementById('resetConfirm').value;

		var ok = true;
		if (!username) {
			setError('resetUsernameError', 'Username wajib diisi');
			ok = false;
		}
		if (password.length < 8) {
			setError('resetPasswordError', 'Kata sandi baru minimal 8 karakter');
			ok = false;
		}
		if (confirm !== password) {
			setError('resetConfirmError', 'Ulangi kata sandi dengan benar');
			ok = false;
		}
		if (!ok) return;

		setLoading(resetBtn, 'Menyimpan...');
		apiFetch('/api/auth/reset', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ username: username, password: password })
		})
			.then(function () {
				document.getElementById('resetForm').reset();
				document.getElementById('loginUsername').value = username;
				showAuth('loginView');
				setError('loginError', 'Kata sandi berhasil diubah. Silakan masuk kembali.', true);
			})
			.catch(function (err) {
				setError('resetError', err.message);
			})
			.finally(function () {
				setLoading(resetBtn, null);
			});
	});

	document.getElementById('logoutBtn').addEventListener('click', function () {
		apiFetch('/api/auth/logout', { method: 'POST' })
			.catch(function () {})
			.finally(function () {
				showAuth('loginView');
			});
	});

	/* ---------- wizard reservasi ---------- */

	var steps = [1, 2, 3, 4].map(function (n) {
		return document.getElementById('step' + n);
	});
	var stepItems = Array.prototype.slice.call(
		document.getElementById('stepper').querySelectorAll('li')
	);
	var btns = ['step1Btn', 'step2Btn', 'step3Btn'].map(function (id) {
		var el = document.getElementById(id);
		el.dataset.label = el.textContent;
		return el;
	});

	var tanggalInput = document.getElementById('tanggal');
	var now = new Date();
	tanggalInput.min =
		now.getFullYear() +
		'-' +
		String(now.getMonth() + 1).padStart(2, '0') +
		'-' +
		String(now.getDate()).padStart(2, '0');

	var kuotaSisa = null;
	var booking = null;

	function showStep(n) {
		steps.forEach(function (s, i) {
			s.classList.toggle('active', i + 1 === n);
		});
		stepItems.forEach(function (li, i) {
			li.classList.toggle('active', i + 1 === n);
			li.classList.toggle('done', i + 1 < n);
		});
		window.scrollTo(0, 0);
	}

	function setQuota(state, text) {
		var out = document.getElementById('kuotaOutput');
		out.className = 'quota quota-' + state;
		out.textContent = text;
	}

	function fetchQuota() {
		var kantor = document.getElementById('kantor').value;
		var tanggal = tanggalInput.value;

		if (!kantor || !tanggal) {
			kuotaSisa = null;
			setQuota('neutral', 'Pilih kantor dan tanggal untuk melihat kuota.');
			return;
		}

		fetch('/api/kuota?kantor=' + encodeURIComponent(kantor) + '&tanggal=' + encodeURIComponent(tanggal))
			.then(function (res) {
				return res.json().then(function (data) {
					if (!res.ok) throw new Error(data.error || 'Gagal memuat kuota');
					return data;
				});
			})
			.then(function (data) {
				kuotaSisa = data.sisa;
				if (data.sisa === 0) {
					setQuota('empty', 'Kuota habis untuk tanggal ini.');
				} else if (data.sisa <= 10) {
					setQuota('low', 'Tersisa ' + data.sisa + ' slot. Segera ajukan permohonan.');
				} else {
					setQuota('ok', 'Tersedia ' + data.sisa + ' slot.');
				}
			})
			.catch(function () {
				kuotaSisa = null;
				setQuota('empty', 'Gagal memuat kuota. Coba lagi.');
			});
	}

	document.getElementById('kantor').addEventListener('change', fetchQuota);
	tanggalInput.addEventListener('change', fetchQuota);

	btns[0].addEventListener('click', function () {
		clearErrors(['kantorError', 'tanggalError']);
		var kantor = document.getElementById('kantor').value;
		var tanggal = tanggalInput.value;

		if (!kantor) {
			setError('kantorError', 'Pilih kantor terlebih dahulu');
			return;
		}
		if (!tanggal) {
			setError('tanggalError', 'Pilih tanggal kunjungan');
			return;
		}
		if (kuotaSisa === 0) {
			setError('tanggalError', 'Kuota pada tanggal ini sudah habis');
			return;
		}

		booking = { kantor: kantor, tanggal: tanggal };
		showStep(2);
	});

	document.getElementById('step2Back').addEventListener('click', function () {
		showStep(1);
	});

	btns[1].addEventListener('click', function () {
		clearErrors(['namaError', 'nikError', 'jenisPassportError', 'masaBerlakuError', 'step2Error']);

		var nama = document.getElementById('nama').value.trim();
		var nik = document.getElementById('nik').value.trim();
		var jenis = document.getElementById('jenisPassport').value;
		var masa = document.getElementById('masaBerlaku').value;
		var percepatan = document.getElementById('percepatan').checked;

		var ok = true;
		if (!nama) {
			setError('namaError', 'Nama lengkap wajib diisi');
			ok = false;
		}
		if (!/^\d{16}$/.test(nik)) {
			setError('nikError', 'NIK harus 16 digit angka');
			ok = false;
		}
		if (!jenis) {
			setError('jenisPassportError', 'Pilih jenis paspor');
			ok = false;
		}
		if (masa !== '5' && masa !== '10') {
			setError('masaBerlakuError', 'Pilih masa berlaku');
			ok = false;
		}
		if (!ok) return;

		setLoading(btns[1], 'Memproses...');
		fetch('/api/reservasi', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({
				nama: nama,
				nik: nik,
				kantor: booking.kantor,
				tanggal: booking.tanggal,
				jenis_paspor: jenis,
				masa_berlaku: masa,
				percepatan: percepatan
			})
		})
			.then(function (res) {
				return res.json().then(function (data) {
					if (!res.ok) throw new Error(data.error || 'Reservasi gagal');
					return data;
				});
			})
			.then(function (data) {
				booking = data;
				showStep(3);
			})
			.catch(function (err) {
				setError('step2Error', err.message);
			})
			.finally(function () {
				setLoading(btns[1], null);
			});
	});

	btns[2].addEventListener('click', function () {
		clearErrors(['fileKtpError', 'fileKkError', 'fileSuppError', 'step3Error']);

		var files = [
			{ input: document.getElementById('fileKtp'), errorId: 'fileKtpError', jenis: 'ktp', label: 'KTP' },
			{ input: document.getElementById('fileKk'), errorId: 'fileKkError', jenis: 'kk', label: 'Kartu Keluarga' },
			{ input: document.getElementById('fileSupp'), errorId: 'fileSuppError', jenis: 'dokumen_pendukung', label: 'Dokumen Pendukung' }
		];

		var validExt = /\.(jpe?g|png|pdf)$/i;
		var maxSize = 2 * 1024 * 1024;
		var ok = true;

		files.forEach(function (f) {
			var file = f.input.files[0];
			if (!file) {
				setError(f.errorId, f.label + ' wajib diunggah');
				ok = false;
				return;
			}
			if (!validExt.test(file.name)) {
				setError(f.errorId, 'Hanya JPG, PNG, atau PDF');
				ok = false;
				return;
			}
			if (file.size > maxSize) {
				setError(f.errorId, 'Ukuran maksimal 2 MB');
				ok = false;
			}
		});
		if (!ok) return;

		setLoading(btns[2], 'Mengunggah...');

		var sequence = Promise.resolve();
		files.forEach(function (f) {
			sequence = sequence.then(function () {
				var fd = new FormData();
				fd.append('kode_booking', booking.kode_booking);
				fd.append('jenis_dokumen', f.jenis);
				fd.append('file', f.input.files[0]);
				return fetch('/api/upload', { method: 'POST', body: fd }).then(function (res) {
					return res.json().then(function (data) {
						if (!res.ok) throw new Error(data.error || 'Gagal unggah ' + f.label);
					});
				});
			});
		});

		sequence
			.then(function () {
				renderSummary();
				showStep(4);
			})
			.catch(function (err) {
				setError('step3Error', err.message);
			})
			.finally(function () {
				setLoading(btns[2], null);
			});
	});

	document.getElementById('step4Print').addEventListener('click', function () {
		window.print();
	});

	function renderSummary() {
		var rows = [
			['Kode booking', booking.kode_booking],
			['Nama pemohon', booking.nama],
			['Kantor imigrasi', booking.kantor],
			['Tanggal kunjungan', booking.tanggal],
			['Jenis paspor', booking.jenis_paspor === 'elektronik' ? 'Elektronik' : 'Biasa'],
			['Masa berlaku', booking.masa_berlaku + ' tahun'],
			['Layanan percepatan', booking.percepatan ? 'Ya (+Rp1.000.000)' : 'Tidak'],
			['Rincian biaya', booking.rincian_biaya],
			['Total biaya', formatRupiah(booking.total)]
		];

		var el = document.getElementById('summary');
		el.textContent = '';

		var table = document.createElement('table');
		table.className = 'summary-table';
		rows.forEach(function (row) {
			var tr = document.createElement('tr');
			var th = document.createElement('th');
			th.textContent = row[0];
			var td = document.createElement('td');
			td.textContent = row[1] == null ? '-' : String(row[1]);
			tr.appendChild(th);
			tr.appendChild(td);
			table.appendChild(tr);
		});
		el.appendChild(table);

		var box = document.createElement('div');
		box.className = 'billing-box';
		var label = document.createElement('p');
		label.className = 'billing-label';
		label.textContent = 'Kode Billing';
		var code = document.createElement('p');
		code.className = 'billing-code';
		code.textContent = booking.kode_billing;
		var total = document.createElement('p');
		total.className = 'billing-total';
		total.textContent = 'Total pembayaran ' + formatRupiah(booking.total) + ' (simulasi)';
		box.appendChild(label);
		box.appendChild(code);
		box.appendChild(total);
		el.appendChild(box);

		var note = document.createElement('p');
		note.className = 'hint';
		note.textContent =
			'Dokumen berhasil diunggah. Datang ke kantor pada tanggal reservasi dan tunjukkan kode booking atau kode billing. Pembayaran hanya simulasi.';
		el.appendChild(note);
	}

	showStep(1);

	apiFetch('/api/auth/me')
		.then(function (user) {
			showApp(user);
		})
		.catch(function () {
			showAuth('loginView');
		});
});
