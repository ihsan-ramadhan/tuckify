import './style.css';
import './app.css';

import {
	GetSchedules, SaveSchedule, StartSchedule, StopSchedule, DeleteSchedule,
	GetRules, SaveRules, ValidateRules,
	RunOrganize,
	GetHistory, UndoRun
} from '../wailsjs/go/main/App';

// state management
let currentTab = 'dashboard';

// DOM elements
const tabBtns = document.querySelectorAll('.tab-btn');
const tabContents = document.querySelectorAll('.tab-content');

const schedulesList = document.getElementById('schedules-list');
const addSchedBtn = document.getElementById('add-sched-btn');
const cancelSchedBtn = document.getElementById('cancel-sched-btn');
const schedFormCard = document.getElementById('sched-form-card');
const schedForm = document.getElementById('sched-form');
const schedFormTitle = document.getElementById('sched-form-title');
const schedOrigName = document.getElementById('sched-orig-name');
const schedName = document.getElementById('sched-name');
const schedFolders = document.getElementById('sched-folders');
const schedCron = document.getElementById('sched-cron');
const schedConfig = document.getElementById('sched-config');

const rulesTextarea = document.getElementById('rules-textarea');
const validateRulesBtn = document.getElementById('validate-rules-btn');
const saveRulesBtn = document.getElementById('save-rules-btn');
const validationAlert = document.getElementById('validation-alert');

const historyRows = document.getElementById('history-rows');

const resultsModal = document.getElementById('results-modal');
const closeModalBtn = document.getElementById('close-modal-btn');
const modalResultsRows = document.getElementById('modal-results-rows');

// tab switcher
tabBtns.forEach(btn => {
	btn.addEventListener('click', () => {
		const target = btn.dataset.tab;
		tabBtns.forEach(b => b.classList.remove('active'));
		tabContents.forEach(c => c.classList.remove('active'));

		btn.classList.add('active');
		document.getElementById(target).classList.add('active');
		currentTab = target;

		if (target === 'dashboard') loadSchedules();
		if (target === 'rules') loadRules();
		if (target === 'history') loadHistory();
	});
});

// modal close
closeModalBtn.addEventListener('click', () => resultsModal.classList.add('hidden'));
globalThis.addEventListener('click', (e) => {
	if (e.target === resultsModal) resultsModal.classList.add('hidden');
});

// dashboard / schedules
async function loadSchedules() {
	try {
		schedulesList.innerHTML = '<div class="loading">Loading schedules...</div>';
		const schedules = await GetSchedules();
		if (!schedules || schedules.length === 0) {
			schedulesList.innerHTML = '<div class="no-data">No active schedules configured.</div>';
			return;
		}

		schedulesList.innerHTML = '';
		schedules.forEach(s => {
			const card = document.createElement('div');
			card.className = 'card schedule-card';

			const statusClass = s.Status === 'active' ? 'status-active' : 'status-inactive';
			const statusLabel = s.Status === 'active' ? 'Running' : 'Stopped';

			// format folders display
			const foldersList = s.Folders.map(f => `<span class="badge badge-secondary">${f}</span>`).join(' ');

			// format last run display
			let lastRunText = 'never';
			if (s.LastRun && !s.LastRun.startsWith('0001')) {
				const d = new Date(s.LastRun);
				lastRunText = d.toLocaleString();
			}

			// format last files display
			let lastFilesText = '';
			if (s.LastFiles > 0) {
				lastFilesText = `<span class="badge badge-success">${s.LastFiles} files organized</span>`;
			} else if (s.LastFiles === 0 && !s.LastRun.startsWith('0001')) {
				lastFilesText = '<span class="badge badge-secondary">0 files organized</span>';
			}

			card.innerHTML = `
				<div class="schedule-header">
					<div>
						<h3 class="schedule-title">${s.Name}</h3>
						<span class="status-indicator ${statusClass}">${statusLabel}</span>
					</div>
					<div class="schedule-actions">
						${s.Status === 'active' ?
							`<button class="btn btn-secondary btn-sm stop-btn" data-name="${s.Name}">Stop</button>` :
							`<button class="btn btn-secondary btn-sm start-btn" data-name="${s.Name}">Start</button>`
						}
						<button class="btn btn-secondary btn-sm edit-btn" data-name="${s.Name}">Edit</button>
						<button class="btn btn-danger btn-sm delete-btn" data-name="${s.Name}">Delete</button>
					</div>
				</div>
				<div class="schedule-body">
					<div class="info-row"><strong>Cron:</strong> <code>${s.Cron}</code></div>
					<div class="info-row"><strong>Service:</strong> <code>${s.Service || 'none'}</code></div>
					<div class="info-row"><strong>Folders:</strong> ${foldersList}</div>
					${s.Config ? `<div class="info-row"><strong>Config:</strong> <code>${s.Config}</code></div>` : ''}
					<div class="info-row last-run-info">
						<strong>Last Run:</strong> <span>${lastRunText} ${lastFilesText}</span>
					</div>
				</div>
				<div class="schedule-footer">
					<button class="btn btn-primary btn-sm run-now-btn" data-name="${s.Name}">Run Now</button>
				</div>
			`;
			schedulesList.appendChild(card);
		});

		// register event listeners
		document.querySelectorAll('.stop-btn').forEach(b => b.addEventListener('click', handleStop));
		document.querySelectorAll('.start-btn').forEach(b => b.addEventListener('click', handleStart));
		document.querySelectorAll('.edit-btn').forEach(b => b.addEventListener('click', handleEdit));
		document.querySelectorAll('.delete-btn').forEach(b => b.addEventListener('click', handleDelete));
		document.querySelectorAll('.run-now-btn').forEach(b => b.addEventListener('click', handleRunNow));

	} catch (err) {
		schedulesList.innerHTML = `<div class="error">Error loading schedules: ${err}</div>`;
	}
}

// schedule management functions
addSchedBtn.addEventListener('click', () => {
	schedFormTitle.textContent = 'New Schedule';
	schedOrigName.value = '';
	schedForm.reset();
	schedFormCard.classList.remove('hidden');
});

cancelSchedBtn.addEventListener('click', () => {
	schedFormCard.classList.add('hidden');
});

schedForm.addEventListener('submit', async (e) => {
	e.preventDefault();
	const name = schedName.value.trim();
	const folders = schedFolders.value.split(',').map(f => f.trim()).filter(Boolean);
	const cron = schedCron.value.trim();
	const configPath = schedConfig.value.trim();

	try {
		await SaveSchedule(name, folders, cron, configPath);
		schedFormCard.classList.add('hidden');
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
		alert(`Error starting schedule: ${err}`);
	}
}

async function handleStop(e) {
	const name = e.target.dataset.name;
	try {
		await StopSchedule(name);
		loadSchedules();
	} catch (err) {
		alert(`Error stopping schedule: ${err}`);
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
			schedFolders.value = s.Folders.join(', ');
			schedCron.value = s.Cron;
			schedConfig.value = s.Config || '';
			schedFormCard.classList.remove('hidden');
		}
	} catch (err) {
		alert(`Error editing schedule: ${err}`);
	}
}

async function handleDelete(e) {
	const name = e.target.dataset.name;
	if (confirm(`Are you sure you want to delete schedule "${name}"?`)) {
		try {
			await DeleteSchedule(name);
			loadSchedules();
		} catch (err) {
			alert(`Error deleting schedule: ${err}`);
		}
	}
}

async function handleRunNow(e) {
	const name = e.target.dataset.name;
	try {
		const schedules = await GetSchedules();
		const s = schedules.find(x => x.Name === name);
		if (!s) return;

		e.target.disabled = true;
		e.target.textContent = 'Running...';

		const resJson = await RunOrganize(s.Folders, false);
		const results = JSON.parse(resJson);

		showResultsModal(results);
		loadSchedules();
	} catch (err) {
		alert(`Error running organizer: ${err}`);
	} finally {
		e.target.disabled = false;
		e.target.textContent = 'Run Now';
	}
}

function showResultsModal(results) {
	modalResultsRows.innerHTML = '';
	if (!results || results.length === 0) {
		modalResultsRows.innerHTML = '<tr><td colspan="4" class="text-center">No files matches rules to organize.</td></tr>';
	} else {
		results.forEach(r => {
			const tr = document.createElement('tr');
			const statusLabel = r.Skipped ? `Skipped: ${r.SkipReason}` : 'Success';
			const statusClass = r.Skipped ? 'text-secondary' : 'text-success';

			tr.innerHTML = `
				<td><code>${r.Source}</code></td>
				<td><span class="badge">${r.Action}</span></td>
				<td><code>${r.Destination || '-'}</code></td>
				<td class="${statusClass}">${statusLabel}</td>
			`;
			modalResultsRows.appendChild(tr);
		});
	}
	resultsModal.classList.remove('hidden');
}

// rules config editor
async function loadRules() {
	try {
		const content = await GetRules();
		rulesTextarea.value = content;
		validationAlert.className = 'alert hidden';
	} catch (err) {
		alert(`Error loading rules: ${err}`);
	}
}

validateRulesBtn.addEventListener('click', async () => {
	const content = rulesTextarea.value;
	try {
		const errMsg = await ValidateRules(content);
		if (errMsg) {
			validationAlert.className = 'alert alert-danger';
			validationAlert.textContent = `Validation Error: ${errMsg}`;
		} else {
			validationAlert.className = 'alert alert-success';
			validationAlert.textContent = 'Configuration is valid!';
		}
	} catch (err) {
		validationAlert.className = 'alert alert-danger';
		validationAlert.textContent = `Error: ${err}`;
	}
});

saveRulesBtn.addEventListener('click', async () => {
	const content = rulesTextarea.value;
	try {
		// validate first
		const errMsg = await ValidateRules(content);
		if (errMsg) {
			alert(`Cannot save. Validation failed:\n${errMsg}`);
			return;
		}

		await SaveRules(content);
		validationAlert.className = 'alert alert-success';
		validationAlert.textContent = 'Config saved successfully!';
	} catch (err) {
		alert(`Error saving rules: ${err}`);
	}
});

// execution history
async function loadHistory() {
	try {
		historyRows.innerHTML = '<tr><td colspan="4" class="text-center">Loading history...</td></tr>';
		const runs = await GetHistory();
		if (!runs || runs.length === 0) {
			historyRows.innerHTML = '<tr><td colspan="4" class="text-center">No execution history found.</td></tr>';
			return;
		}

		// sort by newest
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
						`<button class="btn btn-secondary btn-sm undo-btn" data-id="${r.ID}">Undo</button>` :
						`<span class="text-secondary">none</span>`
					}
				</td>
			`;
			historyRows.appendChild(tr);
		});

		// register undo handler
		document.querySelectorAll('.undo-btn').forEach(b => b.addEventListener('click', handleUndo));

	} catch (err) {
		historyRows.innerHTML = `<tr><td colspan="4" class="text-center text-danger">Error: ${err}</td></tr>`;
	}
}

async function handleUndo(e) {
	const id = Number.parseInt(e.target.dataset.id, 10);
	if (confirm(`Are you sure you want to undo run #${id}? This will revert moved files.`)) {
		try {
			e.target.disabled = true;
			e.target.textContent = 'Undoing...';

			const count = await UndoRun(id);
			alert(`Reverted ${count} items successfully.`);
			loadHistory();
		} catch (err) {
			alert(`Error undoing: ${err}`);
		} finally {
			e.target.disabled = false;
			e.target.textContent = 'Undo';
		}
	}
}

// initial load
loadSchedules();
