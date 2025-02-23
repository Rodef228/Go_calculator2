document.getElementById('calculate').addEventListener('click', async () => {
    const expression = document.getElementById('expression').value;
    if (!expression) {
        alert('Введите выражение!');
        return;
    }

    try {
        const response = await fetch('http://localhost:8080/api/v1/calculate', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ expression }),
        });

        if (response.ok) {
            const data = await response.json();
            alert(`Выражение принято. ID задачи: ${data.id}`);
        } else {
            const errorData = await response.json();
            alert(`Ошибка: ${errorData.error}`);
        }
    } catch (error) {
        alert('Ошибка при отправке запроса: ' + error.message);
    }
});

document.getElementById('check-status').addEventListener('click', async () => {
    const taskId = document.getElementById('task-id').value;
    if (!taskId) {
        alert('Введите ID задачи!');
        return;
    }

    try {
        const response = await fetch(`http://localhost:8080/api/v1/expressions/${taskId}`);
        if (response.ok) {
            const data = await response.json();
            const resultText = document.getElementById('result-text');
            resultText.textContent = `Статус: ${data.expression.status}, Результат: ${data.expression.result}`;
        } else {
            const errorData = await response.json();
            alert(`Ошибка: ${errorData.error}`);
        }
    } catch (error) {
        alert('Ошибка при отправке запроса: ' + error.message);
    }
});

document.getElementById('show-expressions').addEventListener('click', async () => {
    try {
        const response = await fetch('http://localhost:8080/api/v1/expressions');
        if (response.ok) {
            const data = await response.json();
            const expressionsList = document.getElementById('expressions');

            // Проверяем, что data.expressions существует и является массивом
            if (data.expressions && Array.isArray(data.expressions)) {
                expressionsList.innerHTML = data.expressions
                    .map(
                        (expr) => `
                        <li>
                            ID: ${expr.id}, 
                            Статус: ${expr.status}, 
                            Результат: ${expr.result || "еще не готов"}
                        </li>`
                    )
                    .join('');
            } else {
                expressionsList.innerHTML = "<li>Нет данных</li>";
            }
        } else {
            const errorData = await response.json();
            alert(`Ошибка: ${errorData.error}`);
        }
    } catch (error) {
        alert('Ошибка при отправке запроса: ' + error.message);
    }
});


// Получить задачу (GET /internal/task)
document.getElementById('get-task').addEventListener('click', async () => {
    try {
        const response = await fetch('http://localhost:8080/internal/task');
        if (response.ok) {
            const data = await response.json();
            const taskInfo = document.getElementById('task-info');
            taskInfo.textContent = JSON.stringify(data.task, null, 2);
        } else {
            const errorData = await response.json();
            alert(`Ошибка: ${errorData.error}`);
        }
    } catch (error) {
        alert('Ошибка при отправке запроса: ' + error.message);
    }
});

// Отправить результат задачи (POST /internal/task)
document.getElementById('submit-task-result').addEventListener('click', async () => {
    const taskId = document.getElementById('task-result-id').value;
    const result = document.getElementById('task-result-value').value;

    if (!taskId || !result) {
        alert('Введите ID задачи и результат!');
        return;
    }

    try {
        const response = await fetch('http://localhost:8080/internal/task', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                id: taskId,
                result: parseFloat(result),
            }),
        });

        if (response.ok) {
            alert('Результат успешно отправлен!');
        } else {
            const errorData = await response.json();
            alert(`Ошибка: ${errorData.error}`);
        }
    } catch (error) {
        alert('Ошибка при отправке запроса: ' + error.message);
    }
});