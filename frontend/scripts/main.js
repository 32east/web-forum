window.onload = function() {
    const removeAvatarButton = document.getElementById('remove-avatar-button');
    const removeAvatarButton2 = document.getElementById('remove-avatar-button-2');
    const form = document.querySelector('.forum-list');
    const menuButton = document.getElementById('menu-button');
    const menu = document.getElementById('menu');
    const logoutButton = document.getElementById('logout-button');
    const loginForm = document.querySelector('.login-form');
    const registerBtn = document.getElementById('register-btn');
    const loginBtn = document.getElementById('login-btn');
    const sendMessageButton = document.getElementById('send-message');

    function saveRefreshToken(refreshToken) {
        localStorage.setItem('refresh_token', refreshToken);
    }

    function getRefreshToken() {
        return localStorage.getItem('refresh_token');
    }

    setTimeout( function() {
    checkTokenValidity(function() { location.reload(); }, true)
    }, 250);

    function checkTokenValidity(callback, disableCallbackOnFail=false) {
        var accessToken = getCookie('access_token');
        try {
            const decodedToken = atob(accessToken.split('.')[1]);
            const tokenData = JSON.parse(decodedToken);
            const expirationTime = tokenData.exp;
            const currentTime = Math.floor(Date.now() / 1000);

            if (currentTime > expirationTime) {
                // Токен истек, обновляем его
                handleErrorAndRefreshToken({ reason: 'token expired' }, callback);
            } else if (!disableCallbackOnFail) {
                callback();
            }
        } catch (error) {
            console.error('Ошибка декодирования токена:', error);

            if (accessToken==="") {
                handleErrorAndRefreshToken({ reason: 'token expired' }, callback);
                console.log("refreshing...")
            }

            if (!disableCallbackOnFail) {
                callback();
            }
        }
    }

    document.addEventListener('click', (e) => {
        if (e.target.id.endsWith("-a-btn"))  {
            e.preventDefault();
            const id = e.target.id.slice(0, e.target.id.length - ("-a-btn").length);
            const container = document.getElementById(id + "-container");
            container.classList.toggle("object-hide");
        } else if (e.target.hasAttribute('href')) {
            e.preventDefault();
            const url = e.target.href;
            checkTokenValidity(() => {
                window.location.href = url;
            });
        }
    });

    document.addEventListener('click', function(e) {
        const target = e.target;

        // Проверяем, кликнули ли на три точки
        if (target.classList.contains('dots')) {
            const dropdown = target.nextElementSibling;
            dropdown.style.display = dropdown.style.display === 'none' || dropdown.style.display === '' ? 'block' : 'none';
        }

        // Закрываем dropdown, если кликнули вне его
        if (!target.closest('.dropdown') && !target.classList.contains('dots')) {
            const dropdowns = document.querySelectorAll('.dropdown');
            dropdowns.forEach(dropdown => dropdown.style.display = 'none');
        }

        // Проверяем, кликнули ли на "Удалить сообщение"
        if (target.classList.contains('delete-message')) {
            const messageItem = target.closest('li');
            const messageId = messageItem.id.split('-')[0]; // Извлекаем ID сообщения

            fetch('/api/v1/admin/message/delete', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ id: Number(messageId) })
            }).then(response => response.json()).then(response => {
                    if (response.success === true) {
                        location.reload();
                    }
            });

            // Закрываем выпадающий список
            const dropdown = messageItem.querySelector('.dropdown');
            dropdown.style.display = 'none';
        }
    });

    // Функция для обработки ошибок и обновления токенов
    function handleErrorAndRefreshToken(error, callback) {
        if (getRefreshToken() === null) { return }

        if (error.reason === "token is expired"  || error.reason === "token expired" || error.reason === "not authorized") {
            fetch('/api/v1/auth/refresh-token', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ refresh_token: getRefreshToken() })
            })
                .then(response => response.json())
                .then(data => {
                    console.log(data);
                    if (data.success) {
                        // Обновляем токен доступа
                        setCookie('access_token', data.access_token, 1);
                        callback(); // Повторяем первоначальный запрос
                    } else {
                        console.error('Ошибка обновления токена:', data.reason);
                    }
                })
                .catch(error => console.error('Ошибка обновления токена:', error));
        } else {
            console.error('Ошибка:', error);
        }
    }

    // Функция для получения значения cookie
    function getCookie(name) {
        const cookies = document.cookie.split(';');
        for (let i = 0; i < cookies.length; i++) {
            const cookie = cookies[i].trim();
            if (cookie.startsWith(name + '=')) {
                return cookie.substring(name.length + 1);
            }
        }
        return '';
    }

    // Функция для установки cookie
    function setCookie(name, value, days) {
        const expires = new Date();
        expires.setTime(expires.getTime() + (days * 24 * 60 * 60 * 1000));
        document.cookie = name + '=' + value + ';expires=' + expires.toUTCString() + ';path=/';
    }

    // Обработчик для кнопки удаления аватара
    if (removeAvatarButton) {
        removeAvatarButton.addEventListener('click', (e) => {
            e.preventDefault();
            const newValues = {
                avatarRemove: true,
                username: form.querySelector('input[placeholder="Юзернейм"]').value,
                description: form.querySelector('textarea[placeholder="Текст описания"]').value,
                signText: form.querySelector('textarea[placeholder="Текст подписи"]').value,
            };
            const formData = new FormData();
            Object.keys(newValues).forEach(key => formData.append(key, newValues[key]));

            fetch('/api/v1/profile/settings', {
                method: 'POST',
                body: formData,
            })
                .then(response => response.json())
                .then(data => {
                    if (data.success) {
                        location.reload();
                    } else {
                        handleErrorAndRefreshToken(data, () => {
                            fetch('/api/v1/profile/settings', {
                                method: 'POST',
                                body: formData,
                            })
                                .then(response => response.json())
                                .then(data => {
                                    if (data.success) {
                                        location.reload();
                                    }
                                })
                                .catch(error => console.error('Ошибка:', error));
                        });
                    }
                })
                .catch(error => console.error('Ошибка:', error));
        });
    }

    // Обработчик для кнопки сохранения настроек
    if (form) {
        const saveButton = document.getElementById('save-settings');
        const username = form.querySelector('input[placeholder="Юзернейм"]');

        if (username) {
            saveButton.addEventListener('click', (e) => {
                e.preventDefault();
                const newValues = {
                    username: username.value,
                    description: form.querySelector('textarea[placeholder="Текст описания"]').value,
                    sex: document.getElementById("sex").value,
                    avatar: form.querySelector('input[type="file"]').files[0],
                    signText: form.querySelector('textarea[placeholder="Текст подписи"]').value,
                };

                const changedValues = newValues;
                const formData = new FormData();
                Object.keys(changedValues).forEach(key => formData.append(key, changedValues[key]));

                fetch('/api/v1/profile/settings', {
                    method: 'POST',
                    body: formData,
                })
                    .then(response => response.json())
                    .then(data => {
                        if (data.success) {
                            location.reload();
                        } else {
                            handleErrorAndRefreshToken(data, () => {
                                fetch('/api/v1/profile/settings', {
                                    method: 'POST',
                                    body: formData,
                                })
                                    .then(response => response.json())
                                    .then(data => {
                                        if (data.success) {
                                            location.reload();
                                        }
                                    })
                                    .catch(error => console.error('Ошибка:', error));
                            });
                        }
                    })
                    .catch(error => console.error('Ошибка:', error));
            });
        }
    }

    // Обработчик для кнопки создания темы
    const select = document.getElementById('categorys');
    const topicNameInput = document.querySelector('input.left-right-5px');
    const topicMessageInput = document.querySelector('textarea.left-right-5px');
    const saveButton2 = document.getElementById('create-topic-btn');
    if (saveButton2) {
        saveButton2.addEventListener('click', (e) => {
            e.preventDefault();
            const categoryId = select.value;
            const topicName = topicNameInput.value.trim();
            const topicMessage = topicMessageInput.value.trim();
            const data = {
                category_id: Number(categoryId),
                name: topicName,
                message: topicMessage
            };

            fetch('/api/v1/topics/create', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(data)
            })
                .then(response => response.json())
                .then(data => {
                    if (data.success) {
                        window.location.href = data.redirect;
                    } else {
                        handleErrorAndRefreshToken(data, () => {
                            fetch('/api/v1/topics/create', {
                                method: 'POST',
                                headers: {
                                    'Content-Type': 'application/json'
                                },
                                body: JSON.stringify(data)
                            })
                                .then(response => response.json())
                                .then(data => {
                                    if (data.success) {
                                        window.location.href = data.redirect;
                                    }
                                })
                                .catch(error => console.error('Ошибка:', error));
                        });
                    }
                })
                .catch(error => console.error('Ошибка:', error));
        });
    }

    // Обработчик для кнопки создания категории
    const categoryName = document.querySelector('input.left-right-5px');
    const categoryDescription = document.querySelector('textarea.left-right-5px');
    const categoryCreateBtn = document.getElementById('create-category-btn');
    if (categoryCreateBtn) {
        categoryCreateBtn.addEventListener('click', (e) => {
            e.preventDefault();
            const topicName = categoryName.value.trim();
            const topicMessage = categoryDescription.value.trim();
            const data = {
                name: topicName,
                description: topicMessage
            };

            fetch('/api/v1/admin/category/create', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(data)
            })
                .then(response => response.json())
                .then(data => {
                    if (data.success) {
                        window.location.href = "/admin/categories";
                    } else {
                        handleErrorAndRefreshToken(data, () => {
                            fetch('/api/v1/admin/category/create', {
                                method: 'POST',
                                headers: {
                                    'Content-Type': 'application/json'
                                },
                                body: JSON.stringify(data)
                            })
                                .then(response => response.json())
                                .then(data => {
                                    if (data.success) {
                                        window.location.href = "/admin/categories";
                                    }
                                })
                                .catch(error => console.error('Ошибка:', error));
                        });
                    }
                })
                .catch(error => console.error('Ошибка:', error));
        });
    }

    function getCoords(elem) { // crossbrowser version
        var box = elem.getBoundingClientRect();
        var body = document.body;
        var docEl = document.documentElement;
        var scrollTop = window.pageYOffset || docEl.scrollTop || body.scrollTop;
        var scrollLeft = window.pageXOffset || docEl.scrollLeft || body.scrollLeft;
        var clientTop = docEl.clientTop || body.clientTop || 0;
        var clientLeft = docEl.clientLeft || body.clientLeft || 0;
        var top  = box.top +  scrollTop - clientTop;
        var left = box.left + scrollLeft - clientLeft;
        return { top: Math.round(top), left: Math.round(left) };
    }

    var createNewCategorybtn = document.getElementById("create-new-category-btn")
    var createNewCategopry = document.getElementById("create-new-category")
    if (createNewCategorybtn) {
        createNewCategorybtn.addEventListener('click', (e) => {
            e.preventDefault();
            createNewCategopry.classList.toggle("object-hide");
        });
    }

    // Обработчик для кнопки меню
    if (menuButton) {
        menuButton.addEventListener('click', (e) => {
            e.preventDefault();
            menuButton.classList.toggle("account-button-opened");
            menu.classList.toggle("menu-container-open");
        });

        let left = getCoords(menuButton).left + menuButton.offsetWidth - 150;
        let existsTimeout = -1

        window.onresize = function() {
            left = getCoords(menuButton).left + menuButton.offsetWidth - 150;
            menu.style.margin = 'auto ' + left + 'px';
            menu.style.transition = '0s';

            clearTimeout(existsTimeout)

            existsTimeout = setTimeout(function() {
                menu.style.transition = '0.12s';
            }, 100)
        }

        menu.style.margin = 'auto ' + left + 'px';
    }

    // Обработчик для кнопки выхода
    if (logoutButton) {
        logoutButton.addEventListener('click', (e) => {
            e.preventDefault();
            fetch('/api/v1/auth/logout', {
                method: 'POST',
            })
                .then(() => {
                    localStorage.removeItem('refresh_token');
                    location.reload();
                })
                .catch(() => {
                    location.reload();
                });
        });
    }

    // Обработчик для кнопки регистрации
    if (registerBtn) {
        registerBtn.addEventListener('click', (e) => {
            e.preventDefault();
            const login = loginForm.querySelector('input[name="login"]').value;
            const password = loginForm.querySelector('input[name="password"]').value;
            const username = loginForm.querySelector('input[name="username"]').value;
            const email = loginForm.querySelector('input[name="email"]').value;
            const formData = new FormData();
            formData.append('login', login);
            formData.append('password', password);
            formData.append('username', username);
            formData.append('email', email);

            fetch('/api/v1/auth/register', {
                method: 'POST',
                body: formData,
            })
                .then(response => response.json())
                .then(data => {
                    if (data.success) {
                        alert("Вы успешно зарегистрировались! Сейчас мы перенаправим вас на страницу с авторизацией...");
                        setTimeout(() => {
                            window.location.href = "/login";
                        }, 1500);
                    } else {
                        alert(`Ошибка: ${data.reason}`);
                    }
                })
                .catch(error => alert('Произошла ошибка на стороне сервера.'));
        });
    }

    if (loginBtn) {
        loginBtn.addEventListener('click', (e) => {
            e.preventDefault();
            const login = loginForm.querySelector('input[name="login"]').value;
            const password = loginForm.querySelector('input[name="password"]').value;
            const formData = new FormData();
            formData.append('login', login);
            formData.append('password', password);

            fetch('/api/v1/auth/login', {
                method: 'POST',
                body: formData,
            })
                .then(response => response.json())
                .then(data => {
                    if (data.success) {
                        saveRefreshToken(data.refresh_token); // Сохраняем refresh_token в localStorage
                        window.location.href = "/";
                    } else {
                        alert(`Ошибка: ${data.reason}`);
                    }
                })
                .catch(error => alert('Произошла ошибка на стороне сервера.'));
        });
    }


// Обработчик для кнопки отправки сообщения
    if (sendMessageButton) {
        sendMessageButton.addEventListener('click', (e) => {
            e.preventDefault();
            const message = document.getElementById('message-textarea').value;

            // Извлекаем ID топика из URL
            var urlParts = window.location.pathname.split('/');
            var topicId = urlParts[urlParts.length - 1] || urlParts[urlParts.length - 2]; // Получаем ID топика

            fetch('/api/v1/topics/send-message', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json' // Указываем, что данные в формате JSON
                },
                body: JSON.stringify({
                    topic_id: Number(topicId),  // Преобразуем ID топика в число
                    message: message
                }),
            })
                .then(response => response.json())
                .then(data => {
                    if (data.success) {
                        // Перезагружаем страницу, если сообщение успешно отправлено
                        location.reload();
                    } else {
                        handleErrorAndRefreshToken(data, () => {
                            // Повторная отправка сообщения, если была ошибка с токеном
                            fetch('/api/v1/topics/send-message', {
                                method: 'POST',
                                headers: {
                                    'Content-Type': 'application/json'
                                },
                                body: JSON.stringify({
                                    topic_id: Number(topicId),
                                    message: message
                                }),
                            })
                                .then(response => response.json())
                                .then(data => {
                                    if (data.success) {
                                        location.reload();
                                    }
                                })
                                .catch(error => console.error('Ошибка:', error));
                        });
                    }
                })
                .catch(error => console.error('Ошибка:', error));
        });
    }

    // Get all elements with class "save-category-settings"
    const saveButtons = document.querySelectorAll('.save-category-settings');

// Add event listener to each button
    saveButtons.forEach(button => {
        button.addEventListener('click', () => {
            // Get the parent element (the li element)
            const li = button.closest('li');

            // Get the input fields
            const nameInput = li.querySelector('input[type="text"]');
            const descriptionInput = li.querySelector('textarea');

            // Get the ID from the parent element's ID attribute
            const id = li.querySelector('a').id.split('-')[0];

            // Create a JSON payload
            const payload = {
                id: Number(id),
                name: nameInput.value,
                description: descriptionInput.value
            };

            // Send the API request
            fetch('/api/v1/admin/category/edit', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(payload)
            })
                .then(response => response.json())
                .then(data => {
                    if (data.success) {
                        // Reload the page
                        window.location.reload();
                    } else {
                        // Display an error message
                        const errorSpan = li.querySelector('.error-message');
                        if (!errorSpan) {
                            const errorSpan = document.createElement('span');
                            errorSpan.className = 'error-message';
                            li.appendChild(errorSpan);
                        }
                        errorSpan.textContent = 'Error: ' + data.error;
                    }
                })
                .catch(error => {
                    console.error(error);
                });
        });
    });

    var lastLink ;
    document.querySelectorAll('.user-link').forEach(link => {
        link.addEventListener('click', (event) => {
            event.preventDefault();
            lastLink = link;
            const modal = document.getElementById('profile-modal');
            const username = link.getAttribute('data-username');
            const email = link.getAttribute('data-email');
            const sex = link.getAttribute('data-sex');
            const avatar = link.getAttribute('data-avatar');
            const description = link.getAttribute('data-description');
            const signText = link.getAttribute('data-sign-text');

            // Populate the modal with user data
            document.getElementById('username').value = username;
            document.getElementById('email').value = email;
            document.getElementById('sex').value = sex;
            if (avatar === "") {
                document.getElementById('avatar').src = "/./imgs/default-avatar.jpg";
            } else {
                document.getElementById('avatar').src = "/./imgs/avatars/" + avatar;
            }
            document.getElementById('description').value = description;
            document.getElementById('sign-text').value = signText;

            modal.style.display = 'block';
        });
    });

    // Close modal
    if (document.querySelector('.close-button')) {
        document.querySelector('.close-button').addEventListener('click', () => {
            document.getElementById('profile-modal').style.display = 'none';
        });
    }


    // Handle form submission
    if (document.getElementById('profile-form')) {
        // Обработчик для кнопки удаления аватара
        if (removeAvatarButton2) {
            removeAvatarButton2.addEventListener('click', (e) => {
                e.preventDefault();
                const username = lastLink.getAttribute('data-username');
                const description = lastLink.getAttribute('data-description');
                const signText = lastLink.getAttribute('data-sign-text');

                const newValues = {
                    avatarRemove: true,
                    username: username,
                    description: description,
                    signText: signText,
                };
                const formData = new FormData();
                Object.keys(newValues).forEach(key => formData.append(key, newValues[key]));
                formData.append("id", lastLink.getAttribute('data-id'));
                fetch('/api/v1/admin/users/edit', {
                    method: 'POST',
                    body: formData,
                })
                    .then(response => response.json())
                    .then(data => {
                        if (data.success) {
                            location.reload();
                        } else {
                            handleErrorAndRefreshToken(data, () => {
                                fetch('/api/v1/admin/users/edit', {
                                    method: 'POST',
                                    body: formData,
                                })
                                    .then(response => response.json())
                                    .then(data => {
                                        if (data.success) {
                                            location.reload();
                                        }
                                    })
                                    .catch(error => console.error('Ошибка:', error));
                            });
                        }
                    })
                    .catch(error => console.error('Ошибка:', error));
            });
        }

        document.getElementById('profile-form').addEventListener('submit', (event) => {
            event.preventDefault();
            const formData = new FormData(event.target);

            // Here you would typically send the formData to the server
            // using fetch or another method to handle the update

            console.log('Form Data:', Object.fromEntries(formData)); // For debugging
            formData.append("id", lastLink.getAttribute('data-id'));
            fetch('/api/v1/admin/users/edit', {
                method: 'POST',
                body: formData,
            })
                .then(response => response.json())
                .then(data => {
                    if (data.success) {
                        location.reload();
                    } else {
                        handleErrorAndRefreshToken(data, () => {
                            fetch('/api/v1/admin/users/edit', {
                                method: 'POST',
                                body: formData,
                            })
                                .then(response => response.json())
                                .then(data => {
                                    if (data.success) {
                                        location.reload();
                                    }
                                })
                                .catch(error => console.error('Ошибка:', error));
                        });
                    }
                })
                .catch(error => console.error('Ошибка:', error));

            document.getElementById('profile-modal').style.display = 'none'; // Close the modal
        });
    }

    // Get all elements with class "save-category-settings"
    const deleteButtons = document.querySelectorAll('.delete-category-settings');

// Add event listener to each button
    deleteButtons.forEach(button => {
        button.addEventListener('click', () => {
            // Get the parent element (the li element)
            const li = button.closest('li');

            // Get the ID from the parent element's ID attribute
            const id = li.querySelector('a').id.split('-')[0];

            // Create a JSON payload
            const payload = {
                id: Number(id),
            };

            // Send the API request
            fetch('/api/v1/admin/category/delete', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(payload)
            })
                .then(response => response.json())
                .then(data => {
                    if (data.success) {
                        // Reload the page
                        window.location.reload();
                    } else {
                        // Display an error message
                        const errorSpan = li.querySelector('.error-message');
                        if (!errorSpan) {
                            const errorSpan = document.createElement('span');
                            errorSpan.className = 'error-message';
                            li.appendChild(errorSpan);
                        }
                        errorSpan.textContent = 'Error: ' + data.error;
                    }
                })
                .catch(error => {
                    console.error(error);
                });
        });
    });

    if (document.getElementById('username-login-send-query')) {
        const currentUrl = new URL(window.location.href);
        document.getElementById('username-login-send-query').addEventListener('click', () => {
            const inputValue = document.getElementById('search-users-by-username-or-login').value.trim(); // Удаляем лишние пробелы
            if (inputValue) { // Проверяем, что поле не пустое
                currentUrl.searchParams.set('search', inputValue);
                window.location.href = currentUrl.toString();
            }
        });

        // Устанавливаем значение в input, если в URL есть параметр search
        const searchParam = currentUrl.searchParams.get('search');
        if (searchParam) {
            document.getElementById('search-users-by-username-or-login').value = searchParam;
        }
    }

};