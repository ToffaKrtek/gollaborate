document.addEventListener('DOMContentLoaded', () => {
	const status = document.getElementById('status');
	const editor = document.getElementById('code-editor');
	const doc = window.currentDoc || 'untitled';
	
	// Формируем WebSocket URL относительно текущего хоста
	const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
	const wsUrl = `${wsProtocol}//${window.location.host}/ws?doc=${encodeURIComponent(doc)}`;
	
	status.textContent = `Подключение к ${doc}...`;
	
	let ws;
	function connect() {
		ws = new WebSocket(wsUrl);
		
		ws.onopen = () => {
			status.textContent = `● Подключён: ${doc}`;
			status.classList.add('connected');
		};
		
		ws.onclose = () => {
			status.textContent = `✗ Отключён: ${doc} (попытка переподключения...)`;
			status.classList.remove('connected');
			// Авто-переподключение через 2 секунды
			setTimeout(connect, 2000);
		};
		
		ws.onerror = (err) => {
			console.error('WebSocket error:', err);
			status.textContent = '✗ Ошибка соединения';
		};
		
		// При получении сообщения от других пользователей — вставляем в редактор
		ws.onmessage = (event) => {
			// В следующем цикле реализуем правильную синхронизацию
			// Сейчас просто добавляем полученный текст (демо-режим)
			if (typeof event.data === 'string') {
				editor.value += event.data;
			}
		};
	}
	
	connect();
	
	// Отправка изменений: в следующем цикле добавим дебунс и дельты
	// Сейчас — заглушка для теста соединения
});
