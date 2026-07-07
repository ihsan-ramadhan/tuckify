import './style.css';
import './app.css';
import { ICONS } from './icons.js';

import {
	GetSchedules, SaveSchedule, StartSchedule, StopSchedule, DeleteSchedule,
	StartupAll, UnstartupAll, RestartSchedule,
	GetVisualRules, SaveVisualRules,
	RunOrganize,
	GetHistory, UndoRun, DeleteHistoryRun, ClearHistory,
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
const cronFieldsContainer = document.getElementById('cron-fields-container');
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

const confirmModal = document.getElementById('confirm-modal');
const closeConfirmModal = document.getElementById('close-confirm-modal');
const confirmTitle = document.getElementById('confirm-title');
const confirmMessage = document.getElementById('confirm-message');
const confirmOkBtn = document.getElementById('confirm-ok-btn');
const confirmCancelBtn = document.getElementById('confirm-cancel-btn');
let confirmCallback = null;

function showConfirmModal(title, message, okLabel, cb, okClass) {
	confirmTitle.textContent = title;
	confirmMessage.textContent = message;
	confirmOkBtn.textContent = okLabel || 'Delete';
	confirmOkBtn.className = okClass || 'btn btn-danger';
	confirmCallback = cb;
	confirmModal.classList.add('active');
}

closeConfirmModal.addEventListener('click', () => confirmModal.classList.remove('active'));
confirmCancelBtn.addEventListener('click', () => confirmModal.classList.remove('active'));
confirmOkBtn.addEventListener('click', () => {
	confirmModal.classList.remove('active');
	if (confirmCallback) confirmCallback();
	confirmCallback = null;
});
// tab switcher
tabs.forEach(btn => {
	btn.addEventListener('click', () => {
		const target = btn.dataset.tab;
		tabs.forEach(b => b.classList.remove('active'));
		tabPanels.forEach(c => c.classList.remove('active'));

		btn.classList.add('active');
		document.getElementById(target).classList.add('active');
		currentTab = target;

		if (target === 'dashboard') loadDashboard();
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
	if (e.target === confirmModal) confirmModal.classList.remove('active');
});

function updateBrowseBtn(input, btn) {
	const hasValue = input.value.trim().length > 0;
	btn.innerHTML = `${ICONS.folder} ${hasValue ? 'Add More' : 'Browse...'}`;
}

// Folder Pickers (wails Go dialog binding)
browseSchedFolder.addEventListener('click', async () => {
	try {
		const path = await SelectDirectory('Select target folder for schedule');
		if (path) {
			const currentVal = schedFolder.value.trim();
			if (currentVal) {
				schedFolder.value = `${currentVal}, ${path}`;
			} else {
				schedFolder.value = path;
			}
			updateBrowseBtn(schedFolder, browseSchedFolder);
		}
	} catch (err) {
		alert(`Error selecting directory: ${err}`);
	}
});

schedFolder.addEventListener('input', () => updateBrowseBtn(schedFolder, browseSchedFolder));

browseRunFolder.addEventListener('click', async () => {
	try {
		const path = await SelectDirectory('Select folder to organize');
		if (path) {
			const currentVal = runFolder.value.trim();
			if (currentVal) {
				runFolder.value = `${currentVal}, ${path}`;
			} else {
				runFolder.value = path;
			}
			updateBrowseBtn(runFolder, browseRunFolder);
		}
	} catch (err) {
		alert(`Error selecting directory: ${err}`);
	}
});

runFolder.addEventListener('input', () => updateBrowseBtn(runFolder, browseRunFolder));

// Generate custom cron field dropdowns
function buildCronFieldOptions() {
	const minuteOpts = ['*', '0', '5', '10', '15', '20', '25', '30', '35', '40', '45', '50', '55'];
	const hourOpts = ['*', ...Array.from({length: 24}, (_, i) => String(i))];
	const dayOpts = ['*', ...Array.from({length: 31}, (_, i) => String(i + 1))];
	const monthNames = ['Jan','Feb','Mar','Apr','May','Jun','Jul','Aug','Sep','Oct','Nov','Dec'];
	const weekdayNames = ['Sun','Mon','Tue','Wed','Thu','Fri','Sat'];

	const fields = [
		{ key: 'min', label: 'Minute', opts: minuteOpts },
		{ key: 'hour', label: 'Hour', opts: hourOpts },
		{ key: 'day', label: 'Day', opts: dayOpts },
		{ key: 'month', label: 'Month', opts: ['*', ...Array.from({length: 12}, (_, i) => String(i + 1))], labels: ['*', ...monthNames] },
		{ key: 'weekday', label: 'Weekday', opts: ['*', '0','1','2','3','4','5','6'], labels: ['Every','Sun','Mon','Tue','Wed','Thu','Fri','Sat'] },
	];

	cronFieldsContainer.innerHTML = fields.map(f => `
		<div class="cron-field">
			<label>${f.label}</label>
			<select class="form-control cron-${f.key}">
				${f.opts.map((o, i) => `<option value="${o}">${(f.labels || f.opts)[i]}</option>`).join('')}
			</select>
		</div>
	`).join('');

	// Update preview on any change
	cronFieldsContainer.querySelectorAll('select').forEach(s => {
		s.addEventListener('change', updateCronPreview);
	});
}

function getCronFromFields() {
	const min = document.querySelector('.cron-min')?.value || '*';
	const hour = document.querySelector('.cron-hour')?.value || '*';
	const day = document.querySelector('.cron-day')?.value || '*';
	const month = document.querySelector('.cron-month')?.value || '*';
	const weekday = document.querySelector('.cron-weekday')?.value || '*';
	return `${min} ${hour} ${day} ${month} ${weekday}`;
}

function setCronFields(cronExpr) {
	const parts = (cronExpr || '* * * * *').split(' ');
	const setVal = (sel, val) => {
		const el = document.querySelector(sel);
		if (el) el.value = val;
	};
	setVal('.cron-min', parts[0] || '*');
	setVal('.cron-hour', parts[1] || '*');
	setVal('.cron-day', parts[2] || '*');
	setVal('.cron-month', parts[3] || '*');
	setVal('.cron-weekday', parts[4] || '*');
	updateCronPreview();
}

function updateCronPreview() {
	document.getElementById('cron-preview-value').textContent = getCronFromFields();
}

buildCronFieldOptions();

// cron option change helper
schedCron.addEventListener('change', () => {
	if (schedCron.value === 'custom') {
		customCronGroup.classList.remove('hidden');
		updateCronPreview();
	} else {
		customCronGroup.classList.add('hidden');
	}
});

// schedules
async function loadDashboard() {
	try {
		// Load summary stats in parallel
		const [schedules, rules, runs] = await Promise.all([
			GetSchedules(),
			GetVisualRules(),
			GetHistory()
		]);

		// Summary cards
		document.getElementById('stat-rules').textContent = (rules || []).length;
		const activeCount = (schedules || []).filter(s => s.status === 'active').length;
		document.getElementById('stat-schedules').textContent = activeCount;

		// Last run - find the most recent valid run
		let lastRunLabel = 'Never';
		if (runs && runs.length > 0) {
			for (let i = runs.length - 1; i >= 0; i--) {
				const ts = runs[i].timestamp;
				if (ts) {
					let d = new Date(ts);
					if (isNaN(d.getTime())) {
						d = new Date(ts.replace(/(\.\d{3})\d+/, '$1'));
					}
					if (!isNaN(d.getTime()) && d.getFullYear() > 2000) {
						const now = new Date();
						const diffMs = now - d;
						if (!isNaN(diffMs) && diffMs >= 0) {
							const diffMins = Math.floor(diffMs / 60000);
							if (diffMins < 1) lastRunLabel = 'Just now';
							else if (diffMins < 60) lastRunLabel = `${diffMins}m ago`;
							else if (diffMins < 1440) lastRunLabel = `${Math.floor(diffMins / 60)}h ago`;
							else lastRunLabel = `${Math.floor(diffMins / 1440)}d ago`;
						}
						break;
					}
				}
			}
		}
		document.getElementById('stat-last-run').textContent = lastRunLabel;

		// Runs this week
		if (runs && runs.length > 0) {
			const now = new Date();
			const weekAgo = new Date(now);
			weekAgo.setDate(weekAgo.getDate() - 7);
			const weekCount = runs.filter(r => {
				if (!r.timestamp) return false;
				const d = new Date(r.timestamp);
				return !isNaN(d.getTime()) && d >= weekAgo;
			}).length;
			document.getElementById('stat-total-runs').textContent = weekCount;
		} else {
			document.getElementById('stat-total-runs').textContent = '0';
		}

		// Schedule list
		const schedulesList = document.getElementById('schedules-list');
		if (!schedules || schedules.length === 0) {
			schedulesList.innerHTML = '<div class="empty-state">No schedules configured. Click "+ Add Schedule" to create one.</div>';
			return;
		}

		schedulesList.innerHTML = '';
		schedules.forEach(s => {
			const card = document.createElement('div');
			card.className = 'card sched-card';

			const statusClass = s.status === 'active' ? '' : 'inactive';
			const statusLabel = s.status === 'active' ? 'Running' : 'Stopped';

			let lastRunText = 'never';
			if (s.last_run) {
				const d = new Date(s.last_run);
				if (!isNaN(d.getTime()) && d.getFullYear() > 1) {
					lastRunText = d.toLocaleString();
				}
			}

			let lastFilesText = '';
			if (s.last_files > 0) {
				lastFilesText = `<span class="badge badge-success">${s.last_files} files organized</span>`;
			} else if (s.last_files === 0 && s.last_run) {
				const d = new Date(s.last_run);
				if (!isNaN(d.getTime()) && d.getFullYear() > 1) {
					lastFilesText = '<span class="badge">0 files</span>';
				}
			}

			// pretty format cron
			let cronLabel = s.cron;
			if (s.cron === '0 * * * *') cronLabel = 'Every Hour';
			else if (s.cron === '0 */6 * * *') cronLabel = 'Every 6 Hours';
			else if (s.cron === '0 12 * * *') cronLabel = 'Daily at Noon';
			else if (s.cron === '0 0 * * *') cronLabel = 'Daily at Midnight';
			else if (s.cron === '0 0 * * 0') cronLabel = 'Weekly (Sunday)';
			else if (s.cron === '0 0 1 * *') cronLabel = 'Monthly (1st)';

			card.innerHTML = `
				<div>
					<div style="display:flex; justify-content:space-between; align-items:center;">
						<span style="font-size:16px; font-weight:600;">${s.name}</span>
						<span class="status-badge ${statusClass}"><span class="status-dot"></span>${statusLabel}</span>
					</div>
					<div class="sched-meta">
						<div class="meta-item">
							<span class="meta-label">Folder Target</span>
							<span class="meta-val"><code>${(s.folders || []).join(', ')}</code></span>
						</div>
						<div class="meta-item">
							<span class="meta-label">Frequency</span>
							<span class="meta-val">${cronLabel}</span>
						</div>
					</div>
				</div>
				<div style="display:flex; justify-content:space-between; align-items:center; border-top:1px solid var(--border-subtle); padding-top:16px; margin-top:12px;">
					<span style="font-size:12px; color:var(--text-tertiary)">Last: ${lastRunText} ${lastFilesText}</span>
					<div style="display:flex; gap:8px;">
						${s.status === 'active' ?
							`<button class="btn btn-secondary stop-btn" style="padding:4px 8px; font-size:11px;" data-name="${s.name}">Stop</button>` :
							`<button class="btn btn-primary start-btn" style="padding:4px 8px; font-size:11px;" data-name="${s.name}">Start</button>`
						}
						<button class="btn btn-secondary restart-btn" style="padding:4px 8px; font-size:11px;" data-name="${s.name}">Restart</button>
						<button class="btn btn-secondary logs-btn" style="padding:4px 8px; font-size:11px;" data-name="${s.name}">Logs</button>
						<button class="btn btn-secondary edit-btn" style="padding:4px 8px; font-size:11px;" data-name="${s.name}">Edit</button>
						<button class="btn btn-danger delete-btn" style="padding:4px 8px; font-size:11px;" data-name="${s.name}">Delete</button>
					</div>
				</div>
			`;
			schedulesList.appendChild(card);
		});

		document.querySelectorAll('.stop-btn').forEach(b => b.addEventListener('click', handleStop));
		document.querySelectorAll('.start-btn').forEach(b => b.addEventListener('click', handleStart));
		document.querySelectorAll('.restart-btn').forEach(b => b.addEventListener('click', handleRestart));
		document.querySelectorAll('.logs-btn').forEach(b => b.addEventListener('click', handleLogs));
		document.querySelectorAll('.edit-btn').forEach(b => b.addEventListener('click', handleEdit));
		document.querySelectorAll('.delete-btn').forEach(b => b.addEventListener('click', handleDelete));

	} catch (err) {
		document.getElementById('schedules-list').innerHTML = `<div class="alert alert-danger">Error: ${err}</div>`;
	}
}

// open/close schedule modal
addSchedBtn.addEventListener('click', () => {
	schedFormTitle.textContent = 'New Schedule';
	schedOrigName.value = '';
	schedForm.reset();
	customCronGroup.classList.add('hidden');
	updateBrowseBtn(schedFolder, browseSchedFolder);
	schedModal.classList.add('active');
});

closeSchedModal.addEventListener('click', () => schedModal.classList.remove('active'));
cancelSchedBtn.addEventListener('click', () => schedModal.classList.remove('active'));

schedForm.addEventListener('submit', async (e) => {
	e.preventDefault();
	const name = schedName.value.trim();
	const folderVal = schedFolder.value.trim();
	const folders = folderVal.split(',').map(s => s.trim()).filter(s => s !== '');

	let cron = schedCron.value;
	if (cron === 'custom') {
		cron = getCronFromFields();
	}
	const configPath = schedConfig.value.trim();

	try {
		await SaveSchedule(name, folders, cron, configPath);
		schedModal.classList.remove('active');
		loadDashboard();
	} catch (err) {
		alert(`Error saving schedule: ${err}`);
	}
});

async function handleStart(e) {
	const name = e.target.dataset.name;
	try {
		await StartSchedule(name);
		loadDashboard();
	} catch (err) {
		alert(`Error starting: ${err}`);
	}
}

async function handleStop(e) {
	const name = e.target.dataset.name;
	try {
		await StopSchedule(name);
		loadDashboard();
	} catch (err) {
		alert(`Error stopping: ${err}`);
	}
}

async function handleRestart(e) {
	const name = e.target.dataset.name;
	const btn = e.target;
	btn.disabled = true;
	const origHtml = btn.innerHTML;
	btn.innerHTML = '<span class="btn-spinner"></span> Restarting...';
	try {
		await RestartSchedule(name);
		loadDashboard();
	} catch (err) {
		alert(`Error restarting: ${err}`);
	} finally {
		btn.disabled = false;
		btn.innerHTML = origHtml;
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
		const s = schedules.find(x => x.name === name);
		if (s) {
			schedFormTitle.textContent = 'Edit Schedule';
			schedOrigName.value = s.name;
			schedName.value = s.name;
			schedFolder.value = (s.folders || []).join(', ');
			schedConfig.value = s.config || '';

			const presets = ['0 * * * *', '0 */6 * * *', '0 12 * * *', '0 0 * * *', '0 0 * * 0', '0 0 1 * *'];
			if (presets.includes(s.cron)) {
				schedCron.value = s.cron;
				customCronGroup.classList.add('hidden');
			} else {
				schedCron.value = 'custom';
				customCronGroup.classList.remove('hidden');
				setCronFields(s.cron);
			}
			updateBrowseBtn(schedFolder, browseSchedFolder);
			schedModal.classList.add('active');
		}
	} catch (err) {
		alert(`Error loading schedule for edit: ${err}`);
	}
}

async function handleDelete(e) {
	const name = e.target.dataset.name;
	showConfirmModal('Delete Schedule', `Delete schedule "${name}"? This will also stop the service.`, 'Delete', async () => {
		try {
			await DeleteSchedule(name);
			loadDashboard();
		} catch (err) {
			alert(`Error deleting: ${err}`);
		}
	});
}

// run manual
runOrganizeBtn.addEventListener('click', () => triggerRun(false));
runDryBtn.addEventListener('click', () => triggerRun(true));

async function triggerRun(dryRun) {
	const folderVal = runFolder.value.trim();
	if (!folderVal) {
		alert('Please select a target folder first!');
		return;
	}

	const folders = folderVal.split(',').map(s => s.trim()).filter(s => s !== '');
	if (folders.length === 0) {
		alert('Please enter a valid target folder!');
		return;
	}

	const btn = dryRun ? runDryBtn : runOrganizeBtn;
	const origHtml = btn.innerHTML;
	btn.disabled = true;
	btn.innerHTML = '<span class="btn-spinner"></span> Processing...';

	try {
		const results = await RunOrganize(folders, dryRun);
		showResultsModal(results, dryRun);
	} catch (err) {
		alert(`Error running organizer: ${err}`);
	} finally {
		btn.disabled = false;
		btn.innerHTML = origHtml;
	}
}

function showResultsModal(results, dryRun) {
	modalResultsRows.innerHTML = '';
	if (!results || results.length === 0) {
		modalResultsRows.innerHTML = '<tr><td colspan="4" style="text-align:center;">No files match rules to organize.</td></tr>';
	} else {
		results.forEach(r => {
			const tr = document.createElement('tr');
			const isPreview = dryRun && !r.skipped;
			const statusLabel = r.skipped ? `Skipped: ${r.skip_reason}` : (isPreview ? 'Preview' : 'Success');
			const statusClass = r.skipped ? 'text-secondary' : (isPreview ? 'badge' : 'badge badge-success');

			tr.innerHTML = `
				<td><code>${r.source}</code></td>
				<td><span class="badge">${r.action}</span></td>
				<td><code>${r.destination || '-'}</code></td>
				<td><span class="${statusClass}">${statusLabel}</span></td>
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

		const showAdv = rule._showAdvanced ? '' : 'hidden';
		const toggleText = rule._showAdvanced ? 'Hide Advanced Options' : 'Show Advanced Options';

		rcard.innerHTML = `
			<div class="rule-main-row">
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
						<button type="button" class="btn btn-secondary browse-dest-btn" data-idx="${idx}">${ICONS.folder} Browse</button>
						<button type="button" class="btn btn-danger remove-rule-btn" data-idx="${idx}">${ICONS.trash}</button>
					</div>
				</div>
			</div>

			<button type="button" class="rule-advanced-toggle" data-idx="${idx}">${toggleText}</button>

			<div class="rule-advanced-panel ${showAdv}">
				<div class="rule-field">
					<span class="rule-field-label">Filename Patterns (Glob, e.g. *Invoice*, *Report*)</span>
					<input type="text" class="form-control pattern-input" data-idx="${idx}" value="${(rule.filename_patterns || []).join(', ')}" placeholder="e.g. *Invoice*, *Report*">
				</div>
				<div class="rule-field">
					<span class="rule-field-label">Filename Regex (Regex, e.g. ^[0-9]{4}-)</span>
					<input type="text" class="form-control regex-input" data-idx="${idx}" value="${(rule.filename_regex || []).join(', ')}" placeholder="e.g. ^[0-9]{4}-">
				</div>
				<div class="rule-field">
					<span class="rule-field-label">Min Size (e.g. 10KB, 5MB)</span>
					<input type="text" class="form-control minsize-input" data-idx="${idx}" value="${rule.min_size || ''}" placeholder="e.g. 10KB">
				</div>
				<div class="rule-field">
					<span class="rule-field-label">Max Size (e.g. 50MB, 1GB)</span>
					<input type="text" class="form-control maxsize-input" data-idx="${idx}" value="${rule.max_size || ''}" placeholder="e.g. 50MB">
				</div>
				<div class="rule-field">
					<span class="rule-field-label">Min Age (e.g. 24h, 7d)</span>
					<input type="text" class="form-control minage-input" data-idx="${idx}" value="${rule.min_age || ''}" placeholder="e.g. 24h">
				</div>
				<div class="rule-field">
					<span class="rule-field-label">Max Age (e.g. 30d, 365d)</span>
					<input type="text" class="form-control maxage-input" data-idx="${idx}" value="${rule.max_age || ''}" placeholder="e.g. 30d">
				</div>
			</div>
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
	document.querySelectorAll('.rule-advanced-toggle').forEach(btn => {
		btn.addEventListener('click', handleToggleAdvanced);
	});
	document.querySelectorAll('.pattern-input').forEach(input => {
		input.addEventListener('input', handlePatternChange);
	});
	document.querySelectorAll('.regex-input').forEach(input => {
		input.addEventListener('input', handleRegexChange);
	});
	document.querySelectorAll('.minsize-input').forEach(input => {
		input.addEventListener('input', e => { activeRules[Number(e.target.dataset.idx)].min_size = e.target.value.trim(); });
	});
	document.querySelectorAll('.maxsize-input').forEach(input => {
		input.addEventListener('input', e => { activeRules[Number(e.target.dataset.idx)].max_size = e.target.value.trim(); });
	});
	document.querySelectorAll('.minage-input').forEach(input => {
		input.addEventListener('input', e => { activeRules[Number(e.target.dataset.idx)].min_age = e.target.value.trim(); });
	});
	document.querySelectorAll('.maxage-input').forEach(input => {
		input.addEventListener('input', e => { activeRules[Number(e.target.dataset.idx)].max_age = e.target.value.trim(); });
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
	const idx = Number(e.currentTarget.dataset.idx);
	showConfirmModal('Delete Rule', `Remove rule #${idx + 1}? This will discard all its settings.`, 'Remove', () => {
		activeRules.splice(idx, 1);
		renderRules();
	});
}

function handleToggleAdvanced(e) {
	const idx = Number(e.target.dataset.idx);
	activeRules[idx]._showAdvanced = !activeRules[idx]._showAdvanced;
	renderRules();
}

function handlePatternChange(e) {
	const idx = Number(e.target.dataset.idx);
	const val = e.target.value;
	activeRules[idx].filename_patterns = val.split(',').map(s => s.trim()).filter(s => s !== '');
}

function handleRegexChange(e) {
	const idx = Number(e.target.dataset.idx);
	const val = e.target.value;
	activeRules[idx].filename_regex = val.split(',').map(s => s.trim()).filter(s => s !== '');
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

resetRulesBtn.addEventListener('click', () => {
	validationAlert.classList.add('hidden');
	loadRulesBuilder();
});

saveRulesBtn.addEventListener('click', async () => {
	// client-side simple verification
	for (const r of activeRules) {
		const hasExt = r.extensions && r.extensions.length > 0;
		const hasPattern = r.filename_patterns && r.filename_patterns.length > 0;
		const hasRegex = r.filename_regex && r.filename_regex.length > 0;
		const hasDest = r.action === 'delete' || (r.destination && r.destination.trim());

		if (!hasExt && !hasPattern && !hasRegex) {
			alert('Each rule must have at least one match criteria: file extension, filename pattern, or filename regex.');
			return;
		}
		if (!hasDest && r.action !== 'delete') {
			alert('Destination folder is required when action is "Move File"!');
			return;
		}
	}

	let origHtml;
	try {
		origHtml = saveRulesBtn.innerHTML;
		saveRulesBtn.disabled = true;
		saveRulesBtn.innerHTML = '<span class="btn-spinner"></span> Saving...';

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
	} finally {
		saveRulesBtn.disabled = false;
		saveRulesBtn.innerHTML = origHtml || 'Save Rules';
	}
});

// history
async function loadHistory() {
	try {
		historyRows.innerHTML = '<tr><td colspan="4" style="text-align:center;">Loading history...</td></tr>';
		const runs = await GetHistory();

		historyRows.innerHTML = '';

		if (!runs || runs.length === 0) {
			historyRows.innerHTML = '<tr><td colspan="4" style="text-align:center;">No execution history found.</td></tr>';
			return;
		}

		runs.sort((a, b) => (b.id || 0) - (a.id || 0));

		runs.forEach(r => {
			const tr = document.createElement('tr');
			let dateStr = '\u2014';
			const ts = r.timestamp;
			if (ts) {
				let d = new Date(ts);
				if (isNaN(d.getTime())) {
					d = new Date(ts.replace(/(\.\d{3})\d+/, '$1'));
				}
				if (!isNaN(d.getTime()) && d.getFullYear() > 2000) {
					dateStr = d.toLocaleString();
				}
			}

			const foldersText = (r.folders || []).join(', ');
			const entriesArr = r.entries || [];
			const movedCount = entriesArr.filter(e => e.action === 'move' || !e.action).length;

			tr.innerHTML = `
				<td>${dateStr}</td>
				<td><code>${foldersText || '\u2014'}</code></td>
				<td><span class="badge">${movedCount} items</span></td>
				<td>
					<div style="display:flex; gap:6px;">
					${movedCount > 0 ?
						`<button class="btn btn-secondary undo-btn" style="padding:4px 8px; font-size:11px;" data-id="${r.id}">Undo</button>` :
						`<span style="color:var(--text-tertiary)">none</span>`
					}
					<button class="btn btn-danger delete-history-btn" style="padding:4px 8px; font-size:11px;" data-id="${r.id}">${ICONS.trash}</button>
					</div>
				</td>
			`;
			historyRows.appendChild(tr);
		});

		document.querySelectorAll('.undo-btn').forEach(b => b.addEventListener('click', handleUndo));
		document.querySelectorAll('.delete-history-btn').forEach(b => b.addEventListener('click', handleDeleteHistory));

	} catch (err) {
		historyRows.innerHTML = `<tr><td colspan="4" style="text-align:center; color:var(--danger)">Error: ${err}</td></tr>`;
	}
}

async function handleDeleteHistory(e) {
	const id = Number(e.currentTarget.dataset.id);
	showConfirmModal('Delete History', `Delete history entry #${id}? This cannot be undone.`, 'Delete', async () => {
		try {
			await DeleteHistoryRun(id);
			loadHistory();
		} catch (err) {
			alert(`Error deleting history: ${err}`);
		}
	});
}

async function handleUndo(e) {
	const id = Number(e.target.dataset.id);
	const btn = e.target;
	const origHtml = btn.innerHTML;
	showConfirmModal('Undo Run', `Are you sure you want to revert run #${id}? Files will be moved back to their original locations.`, 'Undo', async () => {
		try {
			btn.disabled = true;
			btn.innerHTML = '<span class="btn-spinner"></span> Undoing...';

			const count = await UndoRun(id);
			alert(`Successfully restored ${count} file(s).`);
			loadHistory();
		} catch (err) {
			alert(`Error: ${err}`);
		} finally {
			btn.disabled = false;
			btn.innerHTML = origHtml;
		}
	}, 'btn btn-primary');
}

// Clear History handler
document.getElementById('clear-history-btn').addEventListener('click', () => {
	showConfirmModal('Clear History', 'Delete all history entries? This cannot be undone.', 'Clear All', async () => {
		try {
			await ClearHistory();
			loadHistory();
		} catch (err) {
			alert(`Error clearing history: ${err}`);
		}
	}, 'btn btn-danger');
});

// Startup / Unstartup All handlers
document.getElementById('startup-all-btn').addEventListener('click', async () => {
	const schedules = await GetSchedules();
	if (!schedules || schedules.length === 0) {
		alert('No schedules to start. Click "+ Add Schedule" to create one first.');
		return;
	}
	try {
		await StartupAll();
		loadDashboard();
	} catch (err) {
		alert(`Error starting all: ${err}`);
	}
});

document.getElementById('unstartup-all-btn').addEventListener('click', async () => {
	const schedules = await GetSchedules();
	if (!schedules || schedules.length === 0) {
		alert('No schedules configured. Click "+ Add Schedule" to create one first.');
		return;
	}
	showConfirmModal('Stop All', 'Stop all running services?', 'Stop All', async () => {
		try {
			await UnstartupAll();
			loadDashboard();
		} catch (err) {
			alert(`Error stopping all: ${err}`);
		}
	}, 'btn btn-primary');
});

// initial load
loadDashboard();
