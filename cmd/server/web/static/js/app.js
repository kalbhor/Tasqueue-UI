// API base URL
const API_BASE = '/api';

// State
let currentView = 'dashboard';
let autoRefreshInterval = null;
let currentJobsPage = 0;
let jobsPerPage = 20;
let totalJobs = 0;
let currentJobsStatus = '';
let currentQueue = 'tasqueue:tasks';

// Initialize app
document.addEventListener('DOMContentLoaded', () => {
    initTabs();
    initButtons();
    loadDashboard();
    startAutoRefresh();
});

// Tab navigation
function initTabs() {
    const tabs = document.querySelectorAll('.tab');
    tabs.forEach(tab => {
        tab.addEventListener('click', () => {
            const view = tab.dataset.view;
            switchView(view);
        });
    });
}

function switchView(view) {
    currentView = view;

    // Update tabs
    document.querySelectorAll('.tab').forEach(tab => {
        tab.classList.toggle('active', tab.dataset.view === view);
    });

    // Update views
    document.querySelectorAll('.view').forEach(viewEl => {
        viewEl.classList.toggle('active', viewEl.id === `${view}-view`);
    });

    // Load data for the view
    if (view === 'dashboard') {
        loadDashboard();
    }
}

// Button handlers
function initButtons() {
    document.getElementById('refreshBtn').addEventListener('click', () => {
        if (currentView === 'dashboard') {
            loadDashboard();
        }
    });

    // Jobs
    document.getElementById('loadJobsBtn').addEventListener('click', () => {
        currentJobsPage = 0;
        loadJobs();
    });

    document.getElementById('searchJobBtn').addEventListener('click', searchJob);
    document.getElementById('job-search-input').addEventListener('keypress', (e) => {
        if (e.key === 'Enter') searchJob();
    });

    document.getElementById('jobs-prev-btn').addEventListener('click', () => {
        if (currentJobsPage > 0) {
            currentJobsPage--;
            loadJobs();
        }
    });

    document.getElementById('jobs-next-btn').addEventListener('click', () => {
        const maxPage = Math.ceil(totalJobs / jobsPerPage) - 1;
        if (currentJobsPage < maxPage) {
            currentJobsPage++;
            loadJobs();
        }
    });

    // Chains
    document.getElementById('loadChainBtn').addEventListener('click', loadChain);
    document.getElementById('chain-id-input').addEventListener('keypress', (e) => {
        if (e.key === 'Enter') loadChain();
    });

    // Groups
    document.getElementById('loadGroupBtn').addEventListener('click', loadGroup);
    document.getElementById('group-id-input').addEventListener('keypress', (e) => {
        if (e.key === 'Enter') loadGroup();
    });

    // Modal close
    document.querySelector('.close').addEventListener('click', () => {
        document.getElementById('job-modal').classList.remove('active');
    });
}

// Auto refresh
function startAutoRefresh() {
    autoRefreshInterval = setInterval(() => {
        if (currentView === 'dashboard') {
            loadDashboard();
        }
    }, 3000); // Refresh every 3 seconds
}

// Dashboard
async function loadDashboard() {
    try {
        const response = await fetch(`${API_BASE}/stats`);
        const data = await response.json();

        if (data.error) {
            showError('dashboard-view', data.error);
            return;
        }

        // Update stats
        document.getElementById('stat-pending').textContent = data.total_pending || 0;
        document.getElementById('stat-success').textContent = data.total_success || 0;
        document.getElementById('stat-failed').textContent = data.total_failed || 0;

        // Update tasks list
        const tasksList = document.getElementById('tasks-list');
        if (data.registered_tasks && data.registered_tasks.length > 0) {
            tasksList.innerHTML = data.registered_tasks
                .map(task => `<span class="task-badge">${task}</span>`)
                .join('');
        } else {
            tasksList.innerHTML = '<p class="info">No tasks registered</p>';
        }

        // Update queue stats
        const queueStats = document.getElementById('queue-stats');
        if (data.queue_stats && Object.keys(data.queue_stats).length > 0) {
            queueStats.innerHTML = Object.entries(data.queue_stats)
                .map(([queue, count]) => `<span class="queue-badge">${queue}: ${count}</span>`)
                .join('');
        } else {
            queueStats.innerHTML = '<p class="info">No queue data available</p>';
        }

        updateLastUpdate();
    } catch (error) {
        showError('dashboard-view', 'Failed to load dashboard data: ' + error.message);
    }
}

// Jobs
async function loadJobs() {
    const status = document.getElementById('job-status-filter').value;
    const queue = document.getElementById('queue-filter').value || 'tasqueue:tasks';
    const jobsList = document.getElementById('jobs-list');
    const pagination = document.getElementById('jobs-pagination');

    currentJobsStatus = status;
    currentQueue = queue;

    jobsList.innerHTML = '<p class="loading">Loading...</p>';
    pagination.style.display = 'none';

    try {
        let jobs = [];
        let total = 0;
        let allJobIds = [];

        if (status === '') {
            // "All" - get both successful and failed jobs
            const [successResp, failedResp] = await Promise.all([
                fetch(`${API_BASE}/jobs?status=successful`),
                fetch(`${API_BASE}/jobs?status=failed`)
            ]);

            const successData = await successResp.json();
            const failedData = await failedResp.json();

            if (successData.error) throw new Error(successData.error);
            if (failedData.error) throw new Error(failedData.error);

            allJobIds = [...(successData.job_ids || []), ...(failedData.job_ids || [])];
            total = allJobIds.length;

            // Paginate the job IDs
            const start = currentJobsPage * jobsPerPage;
            const end = start + jobsPerPage;
            const pageJobIds = allJobIds.slice(start, end);

            // Fetch details for paginated job IDs
            if (pageJobIds.length > 0) {
                const jobPromises = pageJobIds.map(id =>
                    fetch(`${API_BASE}/jobs/${id}`).then(r => r.json())
                );
                jobs = await Promise.all(jobPromises);
            }
        } else if (status === 'successful' || status === 'failed') {
            // Load by specific status
            const response = await fetch(`${API_BASE}/jobs?status=${status}`);
            const data = await response.json();

            if (data.error) {
                throw new Error(data.error);
            }

            allJobIds = data.job_ids || [];
            total = allJobIds.length;

            // Paginate the job IDs
            const start = currentJobsPage * jobsPerPage;
            const end = start + jobsPerPage;
            const pageJobIds = allJobIds.slice(start, end);

            // Fetch details for paginated job IDs
            if (pageJobIds.length > 0) {
                const jobPromises = pageJobIds.map(id =>
                    fetch(`${API_BASE}/jobs/${id}`).then(r => r.json())
                );
                jobs = await Promise.all(jobPromises);
            }
        } else {
            // Load pending jobs from queue using pagination
            const offset = currentJobsPage * jobsPerPage;
            const response = await fetch(`${API_BASE}/jobs/pending/${queue}/paginated?offset=${offset}&limit=${jobsPerPage}`);
            const data = await response.json();

            if (data.error) {
                throw new Error(data.error);
            }

            jobs = data.jobs || [];
            total = data.total || 0;
        }

        totalJobs = total;

        if (!jobs || jobs.length === 0) {
            jobsList.innerHTML = '<p class="info">No jobs found</p>';
            pagination.style.display = 'none';
            return;
        }

        jobsList.innerHTML = jobs.map(job => renderJobCard(job)).join('');

        // Attach click handlers
        document.querySelectorAll('.job-card').forEach(card => {
            card.addEventListener('click', () => {
                const jobId = card.dataset.jobId;
                showJobDetail(jobId);
            });
        });

        // Update pagination
        updateJobsPagination();
    } catch (error) {
        showError('jobs-view', 'Failed to load jobs: ' + error.message);
        jobsList.innerHTML = `<p class="error">Failed to load jobs: ${error.message}</p>`;
        pagination.style.display = 'none';
    }
}

function updateJobsPagination() {
    const pagination = document.getElementById('jobs-pagination');
    const pageInfo = document.getElementById('jobs-page-info');
    const prevBtn = document.getElementById('jobs-prev-btn');
    const nextBtn = document.getElementById('jobs-next-btn');

    if (totalJobs === 0) {
        pagination.style.display = 'none';
        return;
    }

    const totalPages = Math.ceil(totalJobs / jobsPerPage);
    const start = currentJobsPage * jobsPerPage + 1;
    const end = Math.min((currentJobsPage + 1) * jobsPerPage, totalJobs);

    pageInfo.textContent = `Showing ${start}-${end} of ${totalJobs} jobs (Page ${currentJobsPage + 1}/${totalPages})`;

    prevBtn.disabled = currentJobsPage === 0;
    nextBtn.disabled = currentJobsPage >= totalPages - 1;

    pagination.style.display = 'flex';
}

async function searchJob() {
    const jobId = document.getElementById('job-search-input').value.trim();
    const jobsList = document.getElementById('jobs-list');
    const pagination = document.getElementById('jobs-pagination');

    if (!jobId) {
        jobsList.innerHTML = '<p class="error">Please enter a Job ID</p>';
        return;
    }

    jobsList.innerHTML = '<p class="loading">Searching...</p>';
    pagination.style.display = 'none';

    try {
        const response = await fetch(`${API_BASE}/jobs/${jobId}`);
        const job = await response.json();

        if (job.error) {
            throw new Error(job.error);
        }

        jobsList.innerHTML = `
            <div class="search-result-info">
                <p><strong>Found job:</strong> ${jobId}</p>
            </div>
            ${renderJobCard(job)}
        `;

        // Attach click handler
        document.querySelector('.job-card').addEventListener('click', () => {
            showJobDetail(job.ID);
        });
    } catch (error) {
        jobsList.innerHTML = `
            <div class="no-results">
                <p class="error">Job not found: ${jobId}</p>
                <p class="info">The job ID may not exist or the job may have been deleted.</p>
            </div>
        `;
    }
}

function renderJobCard(job) {
    const statusClass = `status-${job.Status || job.status || 'queued'}`.toLowerCase();
    const jobId = job.ID || job.id;
    const taskName = job.Job?.Task || job.task || 'Unknown';
    const status = job.Status || job.status || 'queued';
    const queue = job.Queue || job.queue || 'default';
    const retried = job.Retried || 0;
    const maxRetry = job.MaxRetry || 0;

    return `
        <div class="job-card" data-job-id="${jobId}">
            <div class="job-header">
                <span class="job-id">${jobId}</span>
                <span class="status-badge ${statusClass}">${status}</span>
            </div>
            <div class="job-meta">
                <div class="meta-item">
                    <span class="meta-label">Task</span>
                    <span>${taskName}</span>
                </div>
                <div class="meta-item">
                    <span class="meta-label">Queue</span>
                    <span>${queue}</span>
                </div>
                <div class="meta-item">
                    <span class="meta-label">Retries</span>
                    <span>${retried}/${maxRetry}</span>
                </div>
            </div>
        </div>
    `;
}

async function showJobDetail(jobId) {
    const modal = document.getElementById('job-modal');
    const detailDiv = document.getElementById('job-detail');

    modal.classList.add('active');
    detailDiv.innerHTML = '<p class="loading">Loading job details...</p>';

    try {
        const response = await fetch(`${API_BASE}/jobs/${jobId}`);
        const job = await response.json();

        if (job.error) {
            throw new Error(job.error);
        }

        detailDiv.innerHTML = `
            <div class="detail-section">
                <h3>Job Information</h3>
                <div class="detail-content">
                    <strong>ID:</strong> ${job.ID}<br>
                    <strong>Task:</strong> ${job.Job?.Task || 'N/A'}<br>
                    <strong>Status:</strong> <span class="status-badge status-${job.Status.toLowerCase()}">${job.Status}</span><br>
                    <strong>Queue:</strong> ${job.Queue}<br>
                    <strong>Retries:</strong> ${job.Retried}/${job.MaxRetry}<br>
                    <strong>Processed At:</strong> ${job.ProcessedAt || 'N/A'}<br>
                    ${job.PrevErr ? `<strong>Error:</strong> ${job.PrevErr}<br>` : ''}
                </div>
            </div>

            ${job.Job?.Payload ? `
                <div class="detail-section">
                    <h3>Payload</h3>
                    <div class="detail-content">
                        ${formatPayload(job.Job.Payload)}
                    </div>
                </div>
            ` : ''}

            ${job.result_data ? `
                <div class="detail-section">
                    <h3>Result Data</h3>
                    <div class="detail-content">
                        ${formatPayload(job.result_data)}
                    </div>
                </div>
            ` : ''}
        `;
    } catch (error) {
        detailDiv.innerHTML = `<p class="error">Failed to load job details: ${error.message}</p>`;
    }
}

// Chains
async function loadChain() {
    const chainId = document.getElementById('chain-id-input').value.trim();
    const chainsList = document.getElementById('chains-list');

    if (!chainId) {
        chainsList.innerHTML = '<p class="error">Please enter a chain ID</p>';
        return;
    }

    chainsList.innerHTML = '<p class="loading">Loading chain...</p>';

    try {
        const response = await fetch(`${API_BASE}/chains/${chainId}`);
        const chain = await response.json();

        if (chain.error) {
            throw new Error(chain.error);
        }

        chainsList.innerHTML = renderChainDetail(chain);
    } catch (error) {
        chainsList.innerHTML = `<p class="error">Failed to load chain: ${error.message}</p>`;
    }
}

function renderChainDetail(chain) {
    const statusClass = `status-${chain.Status.toLowerCase()}`;

    let jobsHtml = '';
    if (chain.Jobs && chain.Jobs.length > 0) {
        jobsHtml = `
            <div class="chain-progress">
                ${chain.Jobs.map((job, idx) => `
                    <div class="chain-job">
                        <div><strong>${job.Job?.Task || 'Job'}</strong></div>
                        <div><span class="status-badge status-${job.Status.toLowerCase()}">${job.Status}</span></div>
                        <div style="font-size: 11px; margin-top: 5px;">${job.ID}</div>
                    </div>
                    ${idx < chain.Jobs.length - 1 ? '<div class="chain-arrow">→</div>' : ''}
                `).join('')}
            </div>
        `;
    }

    return `
        <div class="chain-card">
            <div class="chain-header">
                <span class="chain-id">${chain.ID}</span>
                <span class="status-badge ${statusClass}">${chain.Status}</span>
            </div>
            <div class="chain-meta">
                <div class="meta-item">
                    <span class="meta-label">Current Job</span>
                    <span>${chain.JobID || 'N/A'}</span>
                </div>
                <div class="meta-item">
                    <span class="meta-label">Completed Jobs</span>
                    <span>${chain.PrevJobs?.length || 0}</span>
                </div>
            </div>
            ${jobsHtml}
        </div>
    `;
}

// Groups
async function loadGroup() {
    const groupId = document.getElementById('group-id-input').value.trim();
    const groupsList = document.getElementById('groups-list');

    if (!groupId) {
        groupsList.innerHTML = '<p class="error">Please enter a group ID</p>';
        return;
    }

    groupsList.innerHTML = '<p class="loading">Loading group...</p>';

    try {
        const response = await fetch(`${API_BASE}/groups/${groupId}`);
        const group = await response.json();

        if (group.error) {
            throw new Error(group.error);
        }

        groupsList.innerHTML = renderGroupDetail(group);
    } catch (error) {
        groupsList.innerHTML = `<p class="error">Failed to load group: ${error.message}</p>`;
    }
}

function renderGroupDetail(group) {
    const statusClass = `status-${group.Status.toLowerCase()}`;

    let jobsHtml = '';
    if (group.Jobs && group.Jobs.length > 0) {
        jobsHtml = `
            <div class="group-jobs-grid">
                ${group.Jobs.map(job => `
                    <div class="group-job-item">
                        <div><strong>${job.Job?.Task || 'Job'}</strong></div>
                        <div style="margin: 8px 0;"><span class="status-badge status-${job.Status.toLowerCase()}">${job.Status}</span></div>
                        <div style="font-size: 11px; color: var(--gray-600);">${job.ID}</div>
                    </div>
                `).join('')}
            </div>
        `;
    }

    return `
        <div class="group-card">
            <div class="group-header">
                <span class="group-id">${group.ID}</span>
                <span class="status-badge ${statusClass}">${group.Status}</span>
            </div>
            <div class="group-meta">
                <div class="meta-item">
                    <span class="meta-label">Total Jobs</span>
                    <span>${Object.keys(group.JobStatus || {}).length}</span>
                </div>
            </div>
            ${jobsHtml}
        </div>
    `;
}

// Utility functions
function formatPayload(data) {
    if (!data) return 'No data';

    try {
        // Try to parse as JSON
        if (typeof data === 'string') {
            const parsed = JSON.parse(atob(data));
            return `<pre>${JSON.stringify(parsed, null, 2)}</pre>`;
        }
        return `<pre>${JSON.stringify(data, null, 2)}</pre>`;
    } catch {
        // If not JSON, display as text
        return typeof data === 'string' ? data : JSON.stringify(data);
    }
}

function updateLastUpdate() {
    const now = new Date();
    document.getElementById('lastUpdate').textContent =
        `Last updated: ${now.toLocaleTimeString()}`;
}

function showError(viewId, message) {
    console.error(message);
}

