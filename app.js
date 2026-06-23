/* =========================================================
   🚀 JAVASCRIPT STATE & VARIABLES
   ========================================================= */
let ws;
let myName = "";
let myToken = "";
let myRole = "user";
let currentReceiver = "";
let currentGroupId = 0;
let typingTimeout;
let dashboardInterval;

let offlineQueue = [];
let currentOffset = 0;
let isLoadingOldMessages = false;
let hasMoreMessages = true;
const msgLimit = 50;
let unreadCounts = {};

/* =========================================================
   🛠️ UI HELPERS & UTILITIES
   ========================================================= */
function formatPreviewText(content) {
    if (!content) return "No messages yet";
    if (content.startsWith("[FILE] ")) return "📎 Attachment";
    if (content.startsWith("📢 [BROADCAST]: ")) return "📢 Announcement";
    return content;
}

function toggleTheme() {
    const body = document.body;
    body.classList.toggle('dark-mode');
    const themeBtn = document.getElementById('themeToggle');
    if (body.classList.contains('dark-mode')) { themeBtn.innerText = '☀️'; } else { themeBtn.innerText = '🌙'; }
}

function toggleEmojiPicker() {
    const picker = document.getElementById('emoji-picker');
    picker.style.display = picker.style.display === 'flex' ? 'none' : 'flex';
}

function insertEmoji(emoji) {
    const input = document.getElementById('msgInput');
    input.value += emoji;
    input.focus();
    document.getElementById('emoji-picker').style.display = 'none';
}

function openLightbox(imgSrc) {
    const lightbox = document.getElementById('lightbox');
    document.getElementById('lightbox-img').src = imgSrc;
    lightbox.style.display = 'flex';
}

function closeLightbox() { document.getElementById('lightbox').style.display = 'none'; }

/* =========================================================
   🔍 SEARCH & EVENT LISTENERS
   ========================================================= */
async function filterUsers() {
    let filter = document.getElementById('userSearchInput').value.toLowerCase();
    let userItems = document.querySelectorAll('#user-list .user-item');

    userItems.forEach(item => {
        let usernameSpan = item.querySelector('.user-item-header span');
        if (usernameSpan) {
            let username = usernameSpan.innerText.replace('👤 ', '').replace(' ⭐ Admin', '').toLowerCase();
            if (username.includes(filter)) { item.style.display = 'flex'; } else { item.style.display = 'none'; }
        }
    });

    if (filter.length >= 2 && myRole !== 'admin') {
        try {
            const res = await fetch(`http://${window.location.host}/search?q=${filter}&token=${myToken}`);
            if (res.ok) {
                const newUsers = await res.json();
                newUsers.forEach(u => {
                    if (!document.getElementById(`user-${u.username}`)) { renderUserInList(u); }
                });
            }
        } catch (e) { console.error("Search error:", e); }
    }
}

// Global Event Listeners setup
document.addEventListener('DOMContentLoaded', () => {
    document.getElementById('user').addEventListener('keydown', function(e) {
        if (e.key === 'ArrowDown' || e.key === 'Enter') {
            e.preventDefault();
            document.getElementById('pass').focus();
        }
    });
    document.getElementById('pass').addEventListener('keydown', function(e) {
        if (e.key === 'Enter') {
            e.preventDefault();
            auth('login');
        }
    });
    document.getElementById('msgInput').addEventListener('keydown', function(e) {
        if (e.key === 'Enter') {
            e.preventDefault();
            sendMsg();
        }
    });
    document.getElementById('groupNameInput').addEventListener('keydown', function(e) {
        if (e.key === 'Enter') {
            e.preventDefault();
            joinGroup();
        }
    });

    document.getElementById('msg-box').addEventListener('scroll', async function() {
        if (this.scrollTop === 0 && hasMoreMessages && !isLoadingOldMessages) {
            if (currentReceiver === "" && currentGroupId === 0) return;
            isLoadingOldMessages = true;
            let prevHeight = this.scrollHeight;
            let url = currentGroupId !== 0 ? `http://${window.location.host}/messages?token=${myToken}&group_id=${currentGroupId}&offset=${currentOffset}` : `http://${window.location.host}/messages?token=${myToken}&with=${currentReceiver}&offset=${currentOffset}`;

            try {
                const res = await fetch(url);
                if (res.ok) {
                    const messages = await res.json();
                    if (messages && messages.length > 0) {
                        if (messages.length < msgLimit) hasMoreMessages = false;
                        currentOffset += messages.length;
                        for (let i = messages.length - 1; i >= 0; i--) {
                            appendMessage(messages[i], "", true);
                            if (currentGroupId === 0 && messages[i].username !== myName && messages[i].status !== "delivered" && messages[i].status !== "read") {
                                if (ws && ws.readyState === WebSocket.OPEN) { ws.send(JSON.stringify({ type: "delivery_ack", id: messages[i].id, receiver_username: messages[i].username })); }
                            }
                        }
                        this.scrollTop = this.scrollHeight - prevHeight;
                    } else { hasMoreMessages = false; }
                }
            } catch (error) { console.error("Error fetching older messages:", error); }
            isLoadingOldMessages = false;
        }
    });
});

function clearLocalChat() {
    const box = document.getElementById('msg-box');
    box.innerHTML = '<p style="text-align:center; color:var(--text-light); margin-top: 50px; font-size: 12px; background: rgba(0,0,0,0.05); padding: 5px; border-radius: 10px; width: fit-content; margin-left: auto; margin-right: auto;">Chat cleared from this device.</p>';
    document.getElementById('msgInput').focus();
}

/* =========================================================
   📢 ADMIN FUNCTIONS (BROADCAST, DASHBOARD, BAN)
   ========================================================= */
function openBroadcastModal() {
    const targetDiv = document.getElementById('broadcastTargets');
    targetDiv.innerHTML = '';

    document.querySelectorAll('#user-list .user-item').forEach(item => {
        const username = item.id.replace('user-', '');
        const displayName = item.querySelector('.user-item-header span').innerText;
        targetDiv.innerHTML += `<label style="display:flex; align-items:center; gap:8px; margin-bottom:8px; padding:5px; background:rgba(0,0,0,0.02); border-radius:5px; cursor:pointer; color:var(--text-dark);"><input type="checkbox" class="broadcast-cb" value="user-${username}"> <span style="font-size:13px;">${displayName}</span></label>`;
    });

    document.querySelectorAll('#group-list .user-item').forEach(item => {
        const groupId = item.id.replace('group-', '');
        const displayName = item.querySelector('.user-item-header span').innerText;
        targetDiv.innerHTML += `<label style="display:flex; align-items:center; gap:8px; margin-bottom:8px; padding:5px; background:rgba(0,0,0,0.02); border-radius:5px; cursor:pointer; color:var(--text-dark);"><input type="checkbox" class="broadcast-cb" value="group-${groupId}"> <span style="font-size:13px;">${displayName}</span></label>`;
    });

    if (targetDiv.innerHTML === '') {
        targetDiv.innerHTML = '<span style="font-size:12px; color:var(--text-light);">No users or groups available to broadcast.</span>';
    }

    document.getElementById('broadcastModal').style.display = 'block';
}

function sendBroadcast() {
    const text = document.getElementById("broadcastText").value;
    if (!text) return;

    const checkboxes = document.querySelectorAll('.broadcast-cb:checked');
    if (checkboxes.length === 0) {
        alert("Please select at least one recipient!");
        return;
    }

    const timeString = new Date().toLocaleTimeString('en-US', { hour: 'numeric', minute: 'numeric', hour12: true });

    checkboxes.forEach(cb => {
        if (cb.value.startsWith('user-')) {
            const targetUser = cb.value.replace('user-', '');
            const payload = {
                id: Date.now().toString() + Math.floor(Math.random() * 1000).toString(),
                type: "chat",
                content: "📢 [BROADCAST]: " + text,
                username: myName,
                receiver_username: targetUser,
                created_at: timeString
            };
            if (ws && ws.readyState === WebSocket.OPEN) { ws.send(JSON.stringify(payload)); }
        } else if (cb.value.startsWith('group-')) {
            const targetGroup = parseInt(cb.value.replace('group-', ''));
            const payload = {
                id: Date.now().toString() + Math.floor(Math.random() * 1000).toString(),
                type: "chat",
                content: "📢 [BROADCAST]: " + text,
                username: myName,
                group_id: targetGroup,
                created_at: timeString
            };
            if (ws && ws.readyState === WebSocket.OPEN) { ws.send(JSON.stringify(payload)); }
        }
    });

    document.getElementById('broadcastModal').style.display = 'none';
    document.getElementById("broadcastText").value = '';
    alert("Broadcast message sent successfully to selected recipients!");
}

function openDashboard() {
    document.getElementById('dashboardModal').style.display = 'block';
    fetchStats();
    dashboardInterval = setInterval(fetchStats, 2000);
}

function closeDashboard() {
    document.getElementById('dashboardModal').style.display = 'none';
    clearInterval(dashboardInterval);
}

async function fetchStats() {
    try {
        const res = await fetch(`http://${window.location.host}/stats?token=${myToken}`);
        if (res.ok) {
            const data = await res.json();
            document.getElementById('stat-goroutines').innerText = data.goroutines;
            document.getElementById('stat-memory').innerText = data.memory_mb + " MB";
            document.getElementById('stat-cpu').innerText = data.cpu_cores + " Cores";
        }
    } catch (e) {}
}

async function banUser(username, e) {
    e.stopPropagation();
    if (!confirm(`Are you sure you want to BAN ${username}? They will be disconnected permanently.`)) return;
    try {
        const res = await fetch(`http://${window.location.host}/ban?token=${myToken}`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username })
        });
        if (res.ok) {
            alert(`User ${username} has been banned.`);
            fetchUsers();
        }
    } catch (err) { console.error(err); }
}

async function unbanUser(username, e) {
    e.stopPropagation();
    if (!confirm(`Are you sure you want to RESTORE ${username}? They will be able to login again.`)) return;
    try {
        const res = await fetch(`http://${window.location.host}/unban?token=${myToken}`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username })
        });
        if (res.ok) {
            alert(`User ${username} has been restored successfully.`);
            fetchUsers();
        }
    } catch (err) { console.error(err); }
}

/* =========================================================
   🌐 HTTP REQUESTS (AUTH, UPLOAD, GROUPS, FETCH)
   ========================================================= */
async function uploadFile() {
    const fileInput = document.getElementById('fileInput');
    if (!fileInput.files.length) return;
    if (currentReceiver === "" && currentGroupId === 0) { alert("Please select a chat first!"); return; }

    const file = fileInput.files[0];
    const formData = new FormData();
    formData.append("file", file);

    try {
        const res = await fetch(`http://${window.location.host}/upload?token=${myToken}`, { method: 'POST', body: formData });
        if (res.ok) {
            const data = await res.json();
            document.getElementById('msgInput').value = `[FILE] ${data.url}`;
            sendMsg();
        } else { alert("Failed to upload file."); }
    } catch (e) { console.error("Upload error:", e); }
    fileInput.value = '';
}

async function auth(type) {
    const u = document.getElementById('user').value;
    const p = document.getElementById('pass').value;
    const s = document.getElementById('secret').value;

    try {
        const res = await fetch(`http://${window.location.host}/${type}`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username: u, password: p, secret: s })
        });

        const data = await res.json();

        if (res.ok) {
            if (type === 'register') {
                document.getElementById('secret').value = '';
                return alert("Registration successful!");
            }

            myName = u;
            myToken = data.token;

            document.getElementById('auth-section').style.display = 'none';
            document.getElementById('chat-layout').style.display = 'flex';
            document.getElementById('my-profile-name').innerText = "User: " + u;

            await fetchUsers();
            await fetchGroups();
            startWS(myToken);
        } else {
            alert(data.message || "Authentication failed");
        }
    } catch (error) { console.error("Auth error:", error); }
}

async function joinGroup() {
    const name = document.getElementById('groupNameInput').value;
    if (!name) return;
    try {
        const res = await fetch(`http://${window.location.host}/groups/join?token=${myToken}`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name: name })
        });
        if (res.ok) {
            document.getElementById('groupNameInput').value = '';
            fetchGroups();
        }
    } catch (e) { console.error("Error joining group:", e); }
}

async function fetchGroups() {
    try {
        const res = await fetch(`http://${window.location.host}/groups?token=${myToken}`);
        if (res.ok) {
            const groups = await res.json();
            const list = document.getElementById('group-list');
            list.innerHTML = '';
            if (groups && groups.length > 0) {
                groups.forEach(g => {
                    const div = document.createElement('div');
                    div.className = 'user-item';
                    div.id = `group-${g.id}`;
                    div.innerHTML = `
                        <div class="user-item-header"><span>👥 ${g.name}</span><div class="indicators" style="display: flex; align-items: center; gap: 5px;"></div></div>
                        <div class="last-msg-preview"><span class="preview-text">${formatPreviewText(g.last_message)}</span><span style="font-size: 10px;">${g.last_time || ""}</span></div>
                    `;
                    div.onclick = () => loadGroupChat(g.id, g.name);
                    list.appendChild(div);
                });
            }
        }
    } catch (error) { console.error("Error fetching groups:", error); }
}

function renderUserInList(u) {
    const div = document.createElement('div');
    div.className = 'user-item';
    div.id = `user-${u.username}`;

    const dot = u.online ? "🟢" : "";
    const adminTag = u.role === "admin" ? " <span style='color:gold; font-size:11px;' title='Administrator'>⭐ Admin</span>" : "";

    let actionBtn = "";
    let userStyle = "";

    if (myRole === "admin" && u.role !== "admin") {
        if (u.is_banned) {
            actionBtn = `<button onclick="unbanUser('${u.username}', event)" style="background: #28a745; padding: 2px 5px; font-size:9px; border-radius: 5px; color: white; border: none; cursor: pointer;" title="Restore User">♻️ Unban</button>`;
            userStyle = "opacity: 0.7;";
        } else {
            actionBtn = `<button onclick="banUser('${u.username}', event)" style="background: #dc3545; padding: 2px 5px; font-size:9px; border-radius: 5px; color: white; border: none; cursor: pointer;" title="Ban User">🚫 Ban</button>`;
        }
    }

    div.innerHTML = `
        <div class="user-item-header" style="${userStyle}">
            <span>👤 ${u.username}${adminTag}</span>
            <div class="indicators" style="display: flex; gap: 5px; align-items: center;">
                ${actionBtn}
                <span class="status-indicator" style="font-size:10px;">${dot}</span>
            </div>
        </div>
        <div class="last-msg-preview" style="${userStyle}"><span class="preview-text">${formatPreviewText(u.last_message)}</span><span style="font-size: 10px;">${u.last_time || ""}</span></div>
    `;

    if (!u.is_banned) {
        div.onclick = () => loadChat(u.username);
        document.getElementById('user-list').appendChild(div);
    } else {
        document.getElementById('banned-list').appendChild(div);
    }
}

async function fetchUsers() {
    try {
        const res = await fetch(`http://${window.location.host}/users?token=${myToken}`);
        if (res.ok) {
            myRole = res.headers.get("X-User-Role") || "user";

            document.getElementById("broadcastBtn").style.display = "inline-block";

            document.getElementById('user-list').innerHTML = '';
            document.getElementById('banned-list').innerHTML = '';

            if (myRole === "admin") {
                document.getElementById("dashboardBtn").style.display = "inline-block";
                document.getElementById("banned-section").style.display = "flex";
            } else {
                document.getElementById("banned-section").style.display = "none";
            }

            const users = await res.json();
            if (users && users.length > 0) { users.forEach(u => renderUserInList(u)); }
        }
    } catch (error) { console.error("Error fetching users:", error); }
}

async function loadGroupChat(groupId, groupName) {
    currentReceiver = "";
    currentGroupId = groupId;
    currentOffset = 0;
    hasMoreMessages = true;
    document.getElementById('chat-title').innerText = `${groupName}`;
    document.getElementById('clearChatBtn').style.display = 'inline-block';
    document.getElementById('msgInput').disabled = false;
    document.getElementById('sendBtn').disabled = false;
    document.getElementById('msgInput').focus();
    document.querySelectorAll('.user-item').forEach(el => el.classList.remove('active'));
    document.getElementById(`group-${groupId}`).classList.add('active');
    unreadCounts[`group-${groupId}`] = 0;
    let badge = document.querySelector(`#group-${groupId} .unread-badge`);
    if (badge) badge.remove();
    hideTypingIndicator();

    try {
        const res = await fetch(`http://${window.location.host}/messages?token=${myToken}&group_id=${groupId}&offset=${currentOffset}`);
        if (res.ok) {
            const messages = await res.json();
            const box = document.getElementById('msg-box');
            box.innerHTML = '';
            if (messages && messages.length > 0) {
                if (messages.length < msgLimit) hasMoreMessages = false;
                currentOffset += messages.length;
                messages.forEach(msg => appendMessage(msg));
            } else {
                box.innerHTML = '<div id="empty-state"><div class="icon">👋</div><p>Start the conversation in this group!</p></div>';
                hasMoreMessages = false;
            }
        }
    } catch (error) { console.error("Error loading group chat:", error); }
}

async function loadChat(receiver) {
    currentReceiver = receiver;
    currentGroupId = 0;
    currentOffset = 0;
    hasMoreMessages = true;
    document.getElementById('chat-title').innerText = `${receiver}`;
    document.getElementById('clearChatBtn').style.display = 'inline-block';
    document.getElementById('msgInput').disabled = false;
    document.getElementById('sendBtn').disabled = false;
    document.getElementById('msgInput').focus();
    document.querySelectorAll('.user-item').forEach(el => el.classList.remove('active'));
    document.getElementById(`user-${receiver}`).classList.add('active');
    unreadCounts[`user-${receiver}`] = 0;
    let badge = document.querySelector(`#user-${receiver} .unread-badge`);
    if (badge) badge.remove();
    hideTypingIndicator();

    if (ws && ws.readyState === WebSocket.OPEN) { ws.send(JSON.stringify({ type: "read_ack", receiver_username: receiver })); }

    try {
        const res = await fetch(`http://${window.location.host}/messages?token=${myToken}&with=${receiver}&offset=${currentOffset}`);
        if (res.ok) {
            const messages = await res.json();
            const box = document.getElementById('msg-box');
            box.innerHTML = '';
            if (messages && messages.length > 0) {
                if (messages.length < msgLimit) hasMoreMessages = false;
                currentOffset += messages.length;
                messages.forEach(msg => {
                    appendMessage(msg);
                    if (msg.username !== myName && msg.status !== "delivered" && msg.status !== "read") {
                        if (ws && ws.readyState === WebSocket.OPEN) { ws.send(JSON.stringify({ type: "delivery_ack", id: msg.id, receiver_username: msg.username })); }
                    }
                });
            } else {
                box.innerHTML = '<div id="empty-state"><div class="icon">👋</div><p>Say hello to ' + receiver + '!</p></div>';
                hasMoreMessages = false;
            }
        }
    } catch (error) { console.error("Error loading chat:", error); }
}

/* =========================================================
   🔌 WEBSOCKET LOGIC (REAL-TIME COMMUNICATION)
   ========================================================= */
function logout() {
    if (ws) { ws.close(); }
    document.getElementById('chat-layout').style.display = 'none';
    document.getElementById('auth-section').style.display = 'block';
    myName = "";
    myToken = "";
    myRole = "user";
    currentReceiver = "";
    currentGroupId = 0;
    document.getElementById('msg-box').innerHTML = '';
    document.getElementById('user').value = '';
    document.getElementById('pass').value = '';
    document.getElementById("broadcastBtn").style.display = "none";
    document.getElementById("dashboardBtn").style.display = "none";
    document.getElementById("banned-section").style.display = "none";
    document.getElementById('secret').style.display = 'none';
    document.getElementById('showAdminBtn').style.display = 'inline-block';
    closeDashboard();
}

function startWS(token) {
    ws = new WebSocket(`ws://${window.location.host}/ws?token=${token}`);
    ws.onopen = () => {
        while (offlineQueue.length > 0) {
            const savedPayload = offlineQueue.shift();
            ws.send(JSON.stringify(savedPayload));
        }
        fetchUsers();
    };
    ws.onmessage = (e) => {
        const data = JSON.parse(e.data);

        if (data.type === "banned") {
            alert("⚠️ Your account has been banned by the Administrator.");
            logout();
            return;
        }

        if (data.type === "status") { updateUserStatus(data.username, data.content); } else if (data.type === "typing") { if (data.username === currentReceiver) { showTypingIndicator(data.username); } } else if (data.type === "server_ack") { const msgDiv = document.getElementById(`msg-${data.id}`); if (msgDiv) { const tickSpan = msgDiv.querySelector('.tick'); if (tickSpan) tickSpan.innerText = "✓"; } } else if (data.type === "delivery_ack") {
            const msgDiv = document.getElementById(`msg-${data.id}`);
            if (msgDiv) {
                const tickSpan = msgDiv.querySelector('.tick');
                if (tickSpan && tickSpan.style.color !== "var(--read-tick)") {
                    tickSpan.innerText = "✓✓";
                    tickSpan.style.color = "#999";
                    tickSpan.style.fontWeight = "normal";
                }
            }
        } else if (data.type === "read_ack") {
            if (data.username === currentReceiver) {
                document.querySelectorAll('.msg.me .tick').forEach(tickSpan => {
                    if (tickSpan.innerText === "✓✓" || tickSpan.innerText === "✓") {
                        tickSpan.innerText = "✓✓";
                        tickSpan.style.color = "var(--read-tick)";
                        tickSpan.style.fontWeight = "bold";
                    }
                });
            }
        } else if (data.type === "chat") {
            if (data.username === myName) return;
            let isUnread = false;
            if (data.group_id && data.group_id !== 0) {
                if (data.group_id !== currentGroupId) isUnread = true;
                updateSidebarPreview(data.username, data.group_id, formatPreviewText(data.content), data.created_at, isUnread);
                if (data.group_id === currentGroupId) { appendMessage(data); }
            } else {
                if (data.username !== currentReceiver) isUnread = true;
                updateSidebarPreview(data.username, data.group_id, formatPreviewText(data.content), data.created_at, isUnread);
                if (!document.getElementById(`user-${data.username}`)) { fetchUsers(); }
                if (data.id) { ws.send(JSON.stringify({ type: "delivery_ack", id: data.id, receiver_username: data.username })); }
                if (data.username === currentReceiver) {
                    hideTypingIndicator();
                    appendMessage(data);
                    ws.send(JSON.stringify({ type: "read_ack", receiver_username: data.username }));
                }
            }
        }
    };
    ws.onclose = () => {
        document.querySelectorAll('.status-indicator').forEach(el => el.innerText = "");
        setTimeout(() => { if (myToken !== "") { startWS(myToken); } }, 3000);
    };
    setInterval(() => { if (ws && ws.readyState === WebSocket.OPEN) { ws.send(JSON.stringify({ type: "ping" })); } }, 30000);
}

function updateUserStatus(username, status) {
    const userDiv = document.getElementById(`user-${username}`);
    if (userDiv) {
        const indicator = userDiv.querySelector('.status-indicator');
        indicator.innerText = status === "online" ? "🟢" : "";
        if (status === "online") {
            let parentList = userDiv.parentElement;
            if (parentList) { parentList.prepend(userDiv); }
        }
    } else if (status === "online") { fetchUsers(); }
}

function updateSidebarPreview(targetUsername, groupId, content, timeStr, incrementUnread = false) {
    let isGroup = groupId !== 0;
    let targetId = isGroup ? `group-${groupId}` : `user-${targetUsername}`;
    let listId = isGroup ? 'group-list' : 'user-list';
    if (incrementUnread) { unreadCounts[targetId] = (unreadCounts[targetId] || 0) + 1; }
    let itemEl = document.getElementById(targetId);
    if (itemEl) {
        let previewEl = itemEl.querySelector('.last-msg-preview');
        if (previewEl) { previewEl.innerHTML = `<span class="preview-text">${content}</span> <span style="font-size: 10px;">${timeStr}</span>`; }
        let indicatorsEl = itemEl.querySelector('.indicators');
        if (indicatorsEl) {
            let badgeEl = indicatorsEl.querySelector('.unread-badge');
            if (unreadCounts[targetId] > 0) {
                if (!badgeEl) {
                    badgeEl = document.createElement('span');
                    badgeEl.className = 'unread-badge';
                    indicatorsEl.appendChild(badgeEl);
                }
                badgeEl.innerText = unreadCounts[targetId];
            } else if (badgeEl) { badgeEl.remove(); }
        }
        let parentList = itemEl.parentElement;
        if (parentList) { parentList.prepend(itemEl); }
    } else {
        if (isGroup) fetchGroups();
        else fetchUsers();
    }
}

function notifyTyping() { if (ws && ws.readyState === WebSocket.OPEN && currentReceiver !== "") { ws.send(JSON.stringify({ type: "typing", receiver_username: currentReceiver })); } }

function showTypingIndicator(username) {
    const ind = document.getElementById('typing-indicator');
    ind.innerText = `${username} is typing...`;
    clearTimeout(typingTimeout);
    typingTimeout = setTimeout(() => { ind.innerText = ""; }, 1500);
}

function hideTypingIndicator() {
    const ind = document.getElementById('typing-indicator');
    if (ind) ind.innerText = "";
    clearTimeout(typingTimeout);
}

function sendMsg() {
    const msgInput = document.getElementById('msgInput');
    if (msgInput.value) {
        if (currentReceiver === "" && currentGroupId === 0) return;
        const uniqueId = Date.now().toString() + Math.floor(Math.random() * 1000).toString();
        const timeString = new Date().toLocaleTimeString('en-US', { hour: 'numeric', minute: 'numeric', hour12: true });

        const payload = { id: uniqueId, type: "chat", content: msgInput.value, username: myName, created_at: timeString };
        if (currentGroupId !== 0) { payload.group_id = currentGroupId; } else { payload.receiver_username = currentReceiver; }

        appendMessage(payload, "🕒");
        updateSidebarPreview(currentReceiver, currentGroupId, "You: " + formatPreviewText(msgInput.value), timeString, false);

        if (ws && ws.readyState === WebSocket.OPEN) { ws.send(JSON.stringify(payload)); } else { offlineQueue.push(payload); }
        msgInput.value = '';
        document.getElementById('emoji-picker').style.display = 'none';
    }
}

function appendMessage(data, tick = "", prepend = false) {
    const box = document.getElementById('msg-box');
    const emptyState = document.getElementById('empty-state');
    if (emptyState) emptyState.remove();
    const placeholder = box.querySelector('p');
    if (placeholder && placeholder.innerText.includes('cleared')) placeholder.remove();

    const div = document.createElement('div');
    if (data.id) div.id = `msg-${data.id}`;
    const timeStr = data.created_at || "";
    let displayContent = data.content;

    if (displayContent.startsWith("[FILE] ")) {
        const fileUrl = displayContent.replace("[FILE] ", "");
        const ext = fileUrl.split('.').pop().toLowerCase();
        if (['png', 'jpg', 'jpeg', 'gif'].includes(ext)) {
            displayContent = `<img src="${fileUrl}" style="max-width: 250px; border-radius: 8px; margin-top: 5px; cursor: pointer; box-shadow: 0 2px 5px rgba(0,0,0,0.2);" onclick="openLightbox('${fileUrl}')">`;
        } else {
            displayContent = `<div style="background: rgba(138, 21, 56, 0.05); padding: 10px; border-radius: 8px; margin-top: 5px; text-align: center; border: 1px solid var(--border-color);"><a href="${fileUrl}" target="_blank" style="color: var(--primary-color); font-weight: bold; text-decoration: none;">📄 Download Attachment</a></div>`;
        }
    }

    if (displayContent.startsWith("📢 [BROADCAST]: ")) {
        const broadcastMsg = displayContent.replace("📢 [BROADCAST]: ", "");
        displayContent = `<div style="background: linear-gradient(135deg, #ffecd2 0%, #fcb69f 100%); padding: 15px; border-radius: 8px; margin-top: 5px; color: #d32f2f; font-weight: bold; border: 2px solid #ffab91; text-align:center;">📢 الإدارة: <br>${broadcastMsg}</div>`;
    }

    if (data.username === myName) {
        div.className = `msg me`;
        let tickIcon = "✓";
        let tickColor = "#999";
        let isBold = "normal";
        if (data.status === "read") {
            tickIcon = "✓✓";
            tickColor = "var(--read-tick)";
            isBold = "bold";
        } else if (data.status === "delivered" || tick === "✓✓") {
            tickIcon = "✓✓";
            tickColor = "#999";
        } else if (tick === "🕒") { tickIcon = "🕒"; }

        div.innerHTML = `<b>${data.username}:</b><br>${displayContent}<div style="display: flex; justify-content: flex-end; align-items: center; margin-top: 5px; gap: 5px; font-size: 11px; color: var(--text-light);"><span>${timeStr}</span><span class="tick" style="font-size:12px; color:${tickColor}; font-weight:${isBold};">${tickIcon}</span></div>`;
    } else {
        div.className = `msg others`;
        div.innerHTML = `<b>${data.username}:</b><br>${displayContent}<div style="display: flex; justify-content: flex-end; margin-top: 5px; font-size: 11px; color: var(--text-light);"><span>${timeStr}</span></div>`;
    }

    if (prepend) { box.prepend(div); } else {
        box.appendChild(div);
        box.scrollTop = box.scrollHeight;
    }
}