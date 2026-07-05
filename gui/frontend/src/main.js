import './style.css';
import './app.css';

import {
	GetSchedules, SaveSchedule, StartSchedule, StopSchedule, DeleteSchedule,
	GetVisualRules, SaveVisualRules,
	RunOrganize,
	GetHistory, UndoRun,
	SelectDirectory, GetLogs,
	GetConflictStrategy, SaveConflictStrategy
} from '../wailsjs/go/main/App';

// state management
let currentTab = 'dashboard';
let activeRules = []; // stores visual rules list

// DOM elements
const tabs = document.querySelectorAll('.tab');
const tabPanels = document.querySelectorAll('.tab-panel');

// Dashboard/Schedules
const schedulesList = document.getElementById('schedules-list');
const addSchedBtn = document.getElementById('add-sched-btn');
const schedModal = document.getElementById('sched-modal');
const closeSchedModal = document.getElementById('close-sched-modal');
const cancelSchedBtn = document.getElementById('cancel-sched-btn');
const schedForm = document.getElementById('sched-form');
const schedFormTitle = document.getElementById('sched-form-title');
const schedOrigName = document.getElementById('sched-orig-name');
const schedName = document.getElementById('sched-name');
const schedFolder = document.getElementById('sched-folder');
const browseSchedFolder = document.getElementById('browse-sched-folder');
const schedCron = document.getElementById('sched-cron');
const customCronGroup = document.getElementById('custom-cron-group');
const schedCustomCron = document.getElementById('sched-custom-cron');
const schedConfig = document.getElementById('sched-config');

// Run Manual
const runFolder = document.getElementById('run-folder');
const browseRunFolder = document.getElementById('browse-run-folder');
const runOrganizeBtn = document.getElementById('run-organize-btn');
const runDryBtn = document.getElementById('run-dry-btn');

// Rules Builder
const rulesList = document.getElementById('rules-list');
const addRuleBtn = document.getElementById('add-rule-btn');
const saveRulesBtn = document.getElementById('save-rules-btn');
const resetRulesBtn = document.getElementById('reset-rules-btn');
const validationAlert = document.getElementById('validation-alert');
const conflictStrategySelect = document.getElementById('conflict-strategy');

// History
const historyRows = document.getElementById('history-rows');
const resultsModal = document.getElementById('results-modal');
const closeModalBtn = document.getElementById('close-modal-btn');
const modalResultsRows = document.getElementById('modal-results-rows');

// Logs Modal
const logsModal = document.getElementById('logs-modal');
const closeLogsModal = document.getElementById('close-logs-modal');
const logsTitle = document.getElementById('logs-title');
const logsContent = document.getElementById('logs-content');

// tab switcher
tabs.forEach(btn => {
	btn.addEventListener('click', () => {
		const target = btn.dataset.tab;
		tabs.forEach(b => b.classList.remove('active'));
		tabPanels.forEach(c => c.classList.remove('active'));

		btn.classList.add('active');
		document.getElementById(target).classList.add('active');
		currentTab = target;

		if (target === 'dashboard') loadSchedules();
		if (target === 'rules-builder') loadRulesBuilder();
		if (target === 'history') loadHistory();
	});
});

// modal close
closeModalBtn.addEventListener('click', () => resultsModal.classList.remove('active'));
closeLogsModal.addEventListener('click', () => logsModal.classList.remove('active'));
globalThis.addEventListener('click', (e) => {
	if (e.target === resultsModal) resultsModal.classList.remove('active');
	if (e.target === logsModal) logsModal.classList.remove('active');
});

// Folder Pickers (wails Go dialog binding)
browseSchedFolder.addEventListener('click', async () => {
	try {
		const path = await SelectDirectory('Select target folder for schedule');
		if (path) {
			schedFolder.value = path;
		}
	} catch (err) {
		alert(`Error selecting directory: ${err}`);
	}
});

browseRunFolder.addEventListener('click', async () => {
	try {
		const path = await SelectDirectory('Select folder to organize');
		if (path) {
			runFolder.value = path;
		}
	} catch (err) {
		alert(`Error selecting directory: ${err}`);
	}
});

// cron option change helper
schedCron.addEventListener('change', () => {
	if (schedCron.value === 'custom') {
		customCronGroup.classList.remove('hidden');
	} else {
		customCronGroup.classList.add('hidden');
	}
});

// schedules
async function loadSchedules() {
	try {
		schedulesList.innerHTML = '<div class="loading">Loading schedules...</div>';
		const schedules = await GetSchedules();
		if (!schedules || schedules.length === 0) {
			schedulesList.innerHTML = '<div class="empty-state">No active schedules configured.</div>';
			return;
		}

		schedulesList.innerHTML = '';
		schedules.forEach(s => {
			const card = document.createElement('div');
			card.className = 'card sched-card';

			const statusClass = s.Status === 'active' ? '' : 'inactive';
			const statusLabel = s.Status === 'active' ? 'Running' : 'Stopped';

			let lastRunText = 'never';
			if (s.LastRun && !s.LastRun.startsWith('0001')) {
				const d = new Date(s.LastRun);
				lastRunText = d.toLocaleString();
			}

			let lastFilesText = '';
			if (s.LastFiles > 0) {
				lastFilesText = `<span class="badge badge-success">${s.LastFiles} files organized</span>`;
			} else if (s.LastFiles === 0 && !s.LastRun.startsWith('0001')) {
				lastFilesText = '<span class="badge">0 files</span>';
			}

			// pretty format cron
			let cronLabel = s.Cron;
			if (s.Cron === '0 * * * *') cronLabel = 'Tiap Jam';
			if (s.Cron === '0 0 * * *') cronLabel = 'Tiap Hari';
			if (s.Cron === '0 0 * * 0') cronLabel = 'Tiap Minggu';

			card.innerHTML = `
				<div>
					<div style="display:flex; justify-content:space-between; align-items:center;">
						<span style="font-size:16px; font-weight:600;">${s.Name}</span>
						<span class="status-badge ${statusClass}"><span class="status-dot"></span>${statusLabel}</span>
					</div>
					<div class="sched-meta">
						<div class="meta-item">
							<span class="meta-label">Folder Target</span>
							<span class="meta-val"><code>${s.Folders.join(', ')}</code></span>
						</div>
						<div class="meta-item">
							<span class="meta-label">Frekuensi</span>
							<span class="meta-val">${cronLabel}</span>
						</div>
					</div>
				</div>
				<div style="display:flex; justify-content:space-between; align-items:center; border-top:1px solid var(--border-subtle); padding-top:16px; margin-top:12px;">
					<span style="font-size:12px; color:var(--text-tertiary)">Last: ${lastRunText} ${lastFilesText}</span>
					<div style="display:flex; gap:8px;">
						${s.Status === 'active' ?
							`<button class="btn btn-secondary stop-btn" style="padding:4px 8px; font-size:11px;" data-name="${s.Name}">Stop</button>` :
							`<button class="btn btn-primary start-btn" style="padding:4px 8px; font-size:11px;" data-name="${s.Name}">Start</button>`
						}
						<button class="btn btn-secondary logs-btn" style="padding:4px 8px; font-size:11px;" data-name="${s.Name}">Logs</button>
						<button class="btn btn-secondary edit-btn" style="padding:4px 8px; font-size:11px;" data-name="${s.Name}">Edit</button>
						<button class="btn btn-danger delete-btn" style="padding:4px 8px; font-size:11px;" data-name="${s.Name}">Delete</button>
					</div>
				</div>
			`;
			schedulesList.appendChild(card);
		});

		document.querySelectorAll('.stop-btn').forEach(b => b.addEventListener('click', handleStop));
		document.querySelectorAll('.start-btn').forEach(b => b.addEventListener('click', handleStart));
		document.querySelectorAll('.logs-btn').forEach(b => b.addEventListener('click', handleLogs));
		document.querySelectorAll('.edit-btn').forEach(b => b.addEventListener('click', handleEdit));
		document.querySelectorAll('.delete-btn').forEach(b => b.addEventListener('click', handleDelete));

	} catch (err) {
		schedulesList.innerHTML = `<div class="alert alert-danger">Error: ${err}</div>`;
	}
}

// open/close schedule modal
addSchedBtn.addEventListener('click', () => {
	schedFormTitle.textContent = 'New Schedule';
	schedOrigName.value = '';
	schedForm.reset();
	customCronGroup.classList.add('hidden');
	schedModal.classList.add('active');
});

closeSchedModal.addEventListener('click', () => schedModal.classList.remove('active'));
cancelSchedBtn.addEventListener('click', () => schedModal.classList.remove('active'));

schedForm.addEventListener('submit', async (e) => {
	e.preventDefault();
	const name = schedName.value.trim();
	const folder = schedFolder.value.trim();
	let cron = schedCron.value;
	if (cron === 'custom') {
		cron = schedCustomCron.value.trim();
	}
	const configPath = schedConfig.value.trim();

	try {
		await SaveSchedule(name, [folder], cron, configPath);
		schedModal.classList.remove('active');
		loadSchedules();
	} catch (err) {
		alert(`Error saving schedule: ${err}`);
	}
});

async function handleStart(e) {
	const name = e.target.dataset.name;
	try {
		await StartSchedule(name);
		loadSchedules();
	} catch (err) {
		alert(`Error starting: ${err}`);
	}
}

async function handleStop(e) {
	const name = e.target.dataset.name;
	try {
		await StopSchedule(name);
		loadSchedules();
	} catch (err) {
		alert(`Error stopping: ${err}`);
	}
}

async function handleLogs(e) {
	const name = e.target.dataset.name;
	logsTitle.textContent = `Logs for: ${name}`;
	logsContent.textContent = 'Fetching logs...';
	logsModal.classList.add('active');
	try {
		const data = await GetLogs(name, 100);
		logsContent.textContent = data || 'No logs available.';
	} catch (err) {
		logsContent.textContent = `Error fetching logs: ${err}`;
	}
}

async function handleEdit(e) {
	const name = e.target.dataset.name;
	try {
		const schedules = await GetSchedules();
		const s = schedules.find(x => x.Name === name);
		if (s) {
			schedFormTitle.textContent = 'Edit Schedule';
			schedOrigName.value = s.Name;
			schedName.value = s.Name;
			schedFolder.value = s.Folders[0] || '';
			schedConfig.value = s.Config || '';
			
			if (['0 * * * *', '0 0 * * *', '0 0 * * 0'].includes(s.Cron)) {
				schedCron.value = s.Cron;
				customCronGroup.classList.add('hidden');
			} else {
				schedCron.value = 'custom';
				schedCustomCron.value = s.Cron;
				customCronGroup.classList.remove('hidden');
			}
			schedModal.classList.add('active');
		}
	} catch (err) {
		alert(`Error loading schedule for edit: ${err}`);
	}
}

async function handleDelete(e) {
	const name = e.target.dataset.name;
	if (confirm(`Hapus schedule "${name}"?`)) {
		try {
			await DeleteSchedule(name);
			loadSchedules();
		} catch (err) {
			alert(`Error deleting: ${err}`);
		}
	}
}

// run manual
runOrganizeBtn.addEventListener('click', () => triggerRun(false));
runDryBtn.addEventListener('click', () => triggerRun(true));

async function triggerRun(dryRun) {
	const folder = runFolder.value.trim();
	if (!folder) {
		alert('Silakan pilih folder target terlebih dahulu!');
		return;
	}

	const btn = dryRun ? runDryBtn : runOrganizeBtn;
	const origText = btn.textContent;
	btn.disabled = true;
	btn.textContent = 'Processing...';

	try {
		const res = await RunOrganize([folder], dryRun);
		showResultsModal(res);
	} catch (err) {
		alert(`Error running organize: ${err}`);
	} finally {
		btn.disabled = false;
		btn.textContent = origText;
	}
}

function showResultsModal(results) {
	modalResultsRows.innerHTML = '';
	if (!results || results.length === 0) {
		modalResultsRows.innerHTML = '<tr><td colspan="4" style="text-align:center;">No files matches rules to organize.</td></tr>';
	} else {
		results.forEach(r => {
			const tr = document.createElement('tr');
			const statusLabel = r.skipped ? `Skipped: ${r.skip_reason}` : 'Success';
			const statusClass = r.skipped ? 'text-secondary' : 'badge badge-success';

			tr.innerHTML = `
				<td><code>${r.source}</code></td>
				<td><span class="badge">${r.action}</span></td>
				<td><code>${r.destination || '-'}</code></td>
				<td><span class="${r.skipped ? '' : statusClass}">${statusLabel}</span></td>
			`;
			modalResultsRows.appendChild(tr);
		});
	}
	resultsModal.classList.add('active');
}

closeModalBtn.addEventListener('click', () => resultsModal.classList.remove('active'));

// rules builder (visual configuration)
async function loadRulesBuilder() {
	try {
		rulesList.innerHTML = '<div class="loading">Loading rules...</div>';
		activeRules = await GetVisualRules();
		const strategy = await GetConflictStrategy();
		conflictStrategySelect.value = strategy;
		renderRules();
	} catch (err) {
		rulesList.innerHTML = `<div class="alert alert-danger">Error loading rules: ${err}</div>`;
	}
}

function renderRules() {
	rulesList.innerHTML = '';
	if (activeRules.length === 0) {
		rulesList.innerHTML = '<div class="empty-state">No rules defined. Click "+ Add New Rule" to create one.</div>';
		return;
	}

	activeRules.forEach((rule, idx) => {
		const rcard = document.createElement('div');
		rcard.className = 'rule-card';

		// tags container (extensions)
		const tagPills = (rule.extensions || []).map(ext => `
			<span class="tag-pill">${ext} <span class="tag-close" data-idx="${idx}" data-val="${ext}">&times;</span></span>
		`).join('');

		rcard.innerHTML = `
			<div class="rule-field">
				<span class="rule-field-label">File Extensions</span>
				<div class="tags-container">
					${tagPills}
					<input type="text" class="tag-input" placeholder="+ add ext..." data-idx="${idx}">
				</div>
			</div>
			<div class="rule-field">
				<span class="rule-field-label">Action</span>
				<select class="form-control action-select" data-idx="${idx}">
					<option value="move" ${rule.action === 'move' ? 'selected' : ''}>Move File</option>
					<option value="delete" ${rule.action === 'delete' ? 'selected' : ''}>Delete File</option>
				</select>
			</div>
			<div class="rule-field">
				<span class="rule-field-label">Destination Folder</span>
				<div class="picker-wrapper">
					<input type="text" class="form-control dest-input" value="${rule.destination || ''}" readonly>
					<button type="button" class="btn btn-secondary browse-dest-btn" data-idx="${idx}">📁 Browse</button>
				</div>
			</div>
			<button type="button" class="btn btn-danger remove-rule-btn" data-idx="${idx}" style="padding: 10px;">&times;</button>
		`;

		rulesList.appendChild(rcard);
	});

	// register rule interaction events
	document.querySelectorAll('.tag-input').forEach(input => {
		input.addEventListener('keydown', handleAddTag);
	});
	document.querySelectorAll('.tag-close').forEach(closeBtn => {
		closeBtn.addEventListener('click', handleRemoveTag);
	});
	document.querySelectorAll('.action-select').forEach(select => {
		select.addEventListener('change', handleActionChange);
	});
	document.querySelectorAll('.browse-dest-btn').forEach(btn => {
		btn.addEventListener('click', handleBrowseDest);
	});
	document.querySelectorAll('.remove-rule-btn').forEach(btn => {
		btn.addEventListener('click', handleRemoveRule);
	});
}

// rule events
function handleAddTag(e) {
	if (e.key === 'Enter') {
		e.preventDefault();
		const val = e.target.value.trim().toLowerCase();
		const idx = Number(e.target.dataset.idx);
		if (val) {
			const formattedVal = val.startsWith('.') ? val : `.${val}`;
			if (!activeRules[idx].extensions) activeRules[idx].extensions = [];
			if (!activeRules[idx].extensions.includes(formattedVal)) {
				activeRules[idx].extensions.push(formattedVal);
				renderRules();
			}
		}
		e.target.value = '';
	}
}

function handleRemoveTag(e) {
	const idx = Number(e.target.dataset.idx);
	const val = e.target.dataset.val;
	activeRules[idx].extensions = activeRules[idx].extensions.filter(ext => ext !== val);
	renderRules();
}

function handleActionChange(e) {
	const idx = Number(e.target.dataset.idx);
	activeRules[idx].action = e.target.value;
}

async function handleBrowseDest(e) {
	const idx = Number(e.target.dataset.idx);
	try {
		const path = await SelectDirectory('Select destination folder');
		if (path) {
			activeRules[idx].destination = path;
			renderRules();
		}
	} catch (err) {
		alert(`Error selecting destination: ${err}`);
	}
}

function handleRemoveRule(e) {
	const idx = Number(e.target.dataset.idx);
	activeRules.splice(idx, 1);
	renderRules();
}

addRuleBtn.addEventListener('click', () => {
	activeRules.push({
		extensions: [],
		filename_patterns: [],
		destination: '',
		action: 'move'
	});
	renderRules();
});

resetRulesBtn.addEventListener('click', loadRulesBuilder);

saveRulesBtn.addEventListener('click', async () => {
	try {
		// client-side simple verification
		for (const r of activeRules) {
			if (!r.extensions || r.extensions.length === 0) {
				alert('Setiap rule wajib memiliki minimal 1 file extension!');
				return;
			}
			if (r.action === 'move' && !r.destination) {
				alert('Destination folder wajib diisi jika action adalah "Move File"!');
				return;
			}
		}

		await SaveVisualRules(activeRules);
		await SaveConflictStrategy(conflictStrategySelect.value);
		validationAlert.className = 'alert alert-success';
		validationAlert.textContent = 'Rules saved successfully!';
		validationAlert.classList.remove('hidden');
		setTimeout(() => validationAlert.classList.add('hidden'), 3000);
	} catch (err) {
		validationAlert.className = 'alert alert-danger';
		validationAlert.textContent = `Error: ${err}`;
		validationAlert.classList.remove('hidden');
	}
});

// history
async function loadHistory() {
	try {
		historyRows.innerHTML = '<tr><td colspan="4" style="text-align:center;">Loading history...</td></tr>';
		const runs = await GetHistory();
		if (!runs || runs.length === 0) {
			historyRows.innerHTML = '<tr><td colspan="4" style="text-align:center;">No execution history found.</td></tr>';
			return;
		}

		runs.sort((a, b) => b.ID - a.ID);

		historyRows.innerHTML = '';
		runs.forEach(r => {
			const tr = document.createElement('tr');
			const d = new Date(r.Timestamp);

			const foldersText = r.Folders.join(', ');
			const movedCount = r.Entries ? r.Entries.length : 0;

			tr.innerHTML = `
				<td>${d.toLocaleString()}</td>
				<td><code>${foldersText}</code></td>
				<td><span class="badge">${movedCount} items</span></td>
				<td>
					${movedCount > 0 ?
						`<button class="btn btn-secondary undo-btn" style="padding:4px 8px; font-size:11px;" data-id="${r.ID}">Undo</button>` :
						`<span style="color:var(--text-tertiary)">none</span>`
					}
				</td>
			`;
			historyRows.appendChild(tr);
		});

		document.querySelectorAll('.undo-btn').forEach(b => b.addEventListener('click', handleUndo));

	} catch (err) {
		historyRows.innerHTML = `<tr><td colspan="4" style="text-align:center; color:var(--danger)">Error: ${err}</td></tr>`;
	}
}

async function handleUndo(e) {
	const id = Number(e.target.dataset.id);
	if (confirm(`Apakah Anda yakin ingin me-revert run #${id}? File-file akan dipindahkan kembali.`)) {
		try {
			e.target.disabled = true;
			e.target.textContent = 'Undoing...';

			const count = await UndoRun(id);
			alert(`Berhasil mengembalikan ${count} file.`);
			loadHistory();
		} catch (err) {
			alert(`Error: ${err}`);
		} finally {
			e.target.disabled = false;
			e.target.textContent = 'Undo';
		}
	}
}

// initial load
loadSchedules();
