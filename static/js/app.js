const app = {
    state: {
        currentPage: 1,
        currentEditId: null,
        searchQuery: '',
        isLoading: false,
        hasMore: true,
        sort: {
            column: 'updated_at',
            direction: 'desc'
        }
    },

    init() {
        this.setupEventListeners();
        this.fetchEntries(true);
    },

    setupEventListeners() {
        const searchInput = document.getElementById('search');
        let debounceTimer;
        const modalOverlay = document.getElementById('modal');
        modalOverlay.addEventListener('click', (e) => {
            if (e.target === modalOverlay) {
                this.hideModal();
            }
        });
        searchInput.addEventListener('input', (e) => {
            clearTimeout(debounceTimer);
            debounceTimer = setTimeout(() => {
                this.state.searchQuery = e.target.value;
                this.fetchEntries(true);
            }, 300);
        });

        const observer = new IntersectionObserver(entries => {
            if (entries[0].isIntersecting && this.state.hasMore && !this.state.isLoading) {
                this.loadMore();
            }
        });
        observer.observe(document.getElementById('loadMore'));
    },

    async fetchEntries(reset = false) {
        if (this.state.isLoading) return;
        
        this.state.isLoading = true;
        if (reset) {
            document.getElementById('entries').innerHTML = '';
            this.state.currentPage = 1;
            this.state.hasMore = true;
        }

        try {
            const response = await axios.get('/api/entries', {
                params: {
                    search: this.state.searchQuery,
                    page: this.state.currentPage,
                    pageSize: 50,
                    sort: this.state.sort.column,
                    order: this.state.sort.direction
                }
            });

            const entries = response.data || [];
            this.processEntries(entries);

        } catch (error) {
            console.error('Fetch error:', error);
            alert(`Error: ${error.response?.data?.error || error.message}`);
        } finally {
            this.state.isLoading = false;
        }
    },

    processEntries(entries) {
        const tbody = document.getElementById('entries');
        this.state.hasMore = entries.length === 50;

        entries.forEach(entry => {
            const row = document.createElement('tr');
            row.className = 'hover:bg-gray-700/50 transition-colors';
            row.innerHTML = `
                <td class="px-6 py-4 whitespace-nowrap text-gray-100 font-medium">${entry.key}</td>
                <td class="px-6 py-4 text-gray-300 value-cell">${entry.value}</td>
                <td class="px-6 py-4 whitespace-nowrap">
                    <button onclick="app.showEditModal(${entry.id})" 
                            class="text-indigo-400 hover:text-indigo-300 mr-2 transition-colors">
                        Edit
                    </button>
                    <button onclick="app.deleteEntry(${entry.id})" 
                            class="text-red-400 hover:text-red-300 transition-colors">
                        Delete
                    </button>
                </td>
            `;
            tbody.appendChild(row);
        });

        this.state.currentPage++;
    },

    sortBy(column) {
        if (this.state.sort.column === column) {
            this.state.sort.direction = this.state.sort.direction === 'asc' ? 'desc' : 'asc';
        } else {
            this.state.sort.column = column;
            this.state.sort.direction = 'asc';
        }

        document.querySelectorAll('.sortable').forEach(el => {
            el.classList.remove('sort-asc', 'sort-desc');
        });
        const header = document.querySelector(`[onclick="app.sortBy('${column}')"]`);
        header.classList.add(`sort-${this.state.sort.direction}`);
        
        this.fetchEntries(true);
    },

    async showEditModal(id) {
        try {
            this.state.currentEditId = id;
            document.body.classList.add('modal-open');
            const response = await axios.get(`/api/entries/${id}`);
            const entry = response.data;
            
            document.getElementById('modalTitle').textContent = 'Edit Entry';
            document.getElementById('key').value = entry.key;
            document.getElementById('value').value = entry.value;
            document.getElementById('modal').classList.remove('hidden');
            
        } catch (error) {
            alert(`Error: ${error.response?.data?.error || error.message}`);
            this.state.currentEditId = null;
        }
    },

    async handleSubmit(event) {
        event.preventDefault();
        const key = document.getElementById('key').value;
        const value = document.getElementById('value').value;

        // Basic validation
        if (!key.trim() || !value.trim()) {
            alert('Please fill in both fields');
            return;
        }

        try {
            const url = this.state.currentEditId 
                ? `/api/entries/${this.state.currentEditId}`
                : '/api/entries';
            
            const method = this.state.currentEditId ? 'put' : 'post';
            
            await axios[method](url, { key, value });
            this.hideModal();
            document.getElementById('key').value = '';
            document.getElementById('value').value = '';
            await this.fetchEntries(true);  // Force refresh the list

        } catch (error) {
            alert(error.response?.data?.error || error.message);
            // Re-open modal if editing
            if (this.state.currentEditId) {
                document.getElementById('modal').classList.remove('hidden');
            }
        }
    },

    async deleteEntry(id) {
        if (confirm('Are you sure you want to delete this entry?')) {
            try {
                await axios.delete(`/api/entries/${id}`);
                await this.fetchEntries(true);
            } catch (error) {
                alert(error.response?.data?.error || error.message);
            }
        }
    },

    async generateDummy() {
        if (confirm('Generate 1000 dummy entries? This might take a few seconds.')) {
            const btn = document.querySelector('[onclick="app.generateDummy()"]');
            try {
                btn.classList.add('btn-loading', 'btn-disabled');
                await axios.post('/api/entries/generate-dummy');
                await this.fetchEntries(true);
                alert('Dummy data generated successfully!');
            } catch (error) {
                alert(`Error: ${error.response?.data?.error || error.message}`);
            } finally {
                btn.classList.remove('btn-loading', 'btn-disabled');
            }
        }
    },

    async truncateDB() {
        if (confirm('WARNING: This will delete ALL entries! Are you sure?')) {
            const btn = document.querySelector('[onclick="app.truncateDB()"]');
            try {
                btn.classList.add('btn-loading', 'btn-disabled');
                await axios.post('/api/entries/truncate');
                await this.fetchEntries(true);
                alert('Database truncated successfully!');
            } catch (error) {
                alert(`Error: ${error.response?.data?.error || error.message}`);
            } finally {
                btn.classList.remove('btn-loading', 'btn-disabled');
            }
        }
    },

    showAddModal() {
        this.state.currentEditId = null;
        document.body.classList.add('modal-open'); // Add blur
        document.getElementById('modalTitle').textContent = 'New Entry';
        document.getElementById('key').value = '';
        document.getElementById('value').value = '';
        document.getElementById('modal').classList.remove('hidden'); // Correct
    },

    hideModal() {
        document.body.classList.remove('modal-open'); // Remove blur
        document.getElementById('modal').classList.add('hidden');
    },

    loadMore() {
        if (!this.state.isLoading) this.fetchEntries();
    }
};

document.addEventListener('DOMContentLoaded', () => app.init());