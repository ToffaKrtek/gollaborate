document.addEventListener('DOMContentLoaded', () => {
	const status = document.getElementById('status');
	const editor = document.getElementById('code-editor');
	const doc = window.currentDoc || 'untitled';
	
	const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
	const wsUrl = `${wsProtocol}//${window.location.host}/ws?doc=${encodeURIComponent(doc)}`;
	
	let ws;
	let reconnectTimer;
	let sendTimer;
	
	function connect() {
		status.textContent = `Подключение к ${doc}...`;
		ws = new WebSocket(wsUrl);
		
		ws.onopen = () => {
			status.textContent = `● Подключён: ${doc}`;
			status.classList.add('connected');
		};
		
		ws.onclose = () => {
			status.textContent = `✗ Отключён: ${doc}`;
			status.classList.remove('connected');
			reconnectTimer = setTimeout(connect, 2000);
		};
		
		ws.onerror = () => {
			status.textContent = '✗ Ошибка соединения';
		};
		
		ws.onmessage = (event) => {
			if (typeof event.data === 'string') {
				// Простая замена: в следующих циклах реализуем операционные преобразования
				editor.value = event.data;
			}
		};
	}
	
	connect();
	
	editor.addEventListener('input', () => {
		if (ws?.readyState !== WebSocket.OPEN) return;
		
		// Отменяем предыдущий таймер
		if (sendTimer) clearTimeout(sendTimer);
		
		// Отправляем с небольшой задержкой, чтобы не спамить
		sendTimer = setTimeout(() => {
			ws.send(editor.value);
		}, 50);
	});
	
	// Очистка при выгрузке
	window.addEventListener('beforeunload', () => {
		if (sendTimer) clearTimeout(sendTimer);
		if (reconnectTimer) clearTimeout(reconnectTimer);
		if (ws) ws.close();
	});
});
