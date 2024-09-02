
window.onload = function() {
    const remove_avatar_button = document.getElementById('remove-avatar-button');

    if (remove_avatar_button) {
        remove_avatar_button.addEventListener('click', (e) => {
            e.preventDefault();

            const newValues = {
                avatarRemove: true,
            };
            const formData = new FormData();

            if (Object.keys(newValues).length > 0) {
                Object.keys(newValues).forEach((key) => {
                    formData.append(key, newValues[key]);
                });
            }

            fetch('/api/profile/settings', {
                method: 'POST',
                body: formData,
            }).then(response => response.json())
                .then((data) => {
                    if (data.success === true) {
                        location.reload()
                    }
                }).catch((error) => {
                console.log(error);
            });
        });
    }

    const form = document.querySelector('.forum-list');

    if (form) {
        const saveButton = document.getElementById('save-settings');
        const username = form.querySelector('input[placeholder="Юзернейм"]');

        if (username) {
            const originalValues = {
                username: form.querySelector('input[placeholder="Юзернейм"]').value,
                description: form.querySelector('input[placeholder="Описание"]').value,
                avatar: null,
                signText: form.querySelector('textarea[placeholder="Текст подписи"]').value,
            };

            saveButton.addEventListener('click', (e) => {
                e.preventDefault();

                const newValues = {
                    username: form.querySelector('input[placeholder="Юзернейм"]').value,
                    description: form.querySelector('input[placeholder="Описание"]').value,
                    avatar: form.querySelector('input[type="file"]').files[0],
                    signText: form.querySelector('textarea[placeholder="Текст подписи"]').value,
                };

                const changedValues = newValues;

                if (Object.keys(changedValues).length > 0) {
                    const formData = new FormData();

                    Object.keys(changedValues).forEach((key) => {
                        formData.append(key, changedValues[key]);
                    });

                    fetch('/api/profile/settings', {
                        method: 'POST',
                        body: formData,
                    })
                        .then((response) => response.json())
                        .then(function (data) {

                            if (!data.success) {
                                return
                            }

                            location.reload()
                        })
                        .catch((error) => console.log(error));
                }
            });
        }
    }

    const select = document.getElementById('categorys');
    const topicNameInput = document.querySelector('input.left-right-5px');
    const topicMessageInput = document.querySelector('textarea.left-right-5px');
    const saveButton2 = document.getElementById('save-settings');
    if (saveButton2) {
        saveButton2.addEventListener('click', (e) => {
            e.preventDefault();

            const categoryId = select.value;
            const topicName = topicNameInput.value.trim();
            const topicMessage = topicMessageInput.value.trim();

            const data = {
                category_id: categoryId,
                name: topicName,
                message: topicMessage
            };

            fetch('/api/topics/create', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(data)
            })
                .then(response => response.json())
                .then((data) => {
                    if (data.success === true) {
                        window.location.href = data.redirect;
                    } else {
                        alert(`Ошибка: ${data.reason}`);
                    }
                })
                .catch((error) => {
                    console.error('Ошибка:', error);
                });
        });
    }

    const menu_button = document.getElementById('menu-button');

    if (menu_button) {
        menu_button.addEventListener('click', (e) => {
            e.preventDefault();

            const menu = document.getElementById("menu");
            if (menu.classList.contains("menu-container-open")) {
                menu_button.classList.remove("account-button-opened")
                menu.classList.remove("menu-container-open")
            } else {
                menu_button.classList.add("account-button-opened")
                menu.classList.add("menu-container-open")
            }
        })
    }

    const logout_button = document.getElementById('logout-button');

    if (logout_button) {
        logout_button.addEventListener('click', (e) => {
            e.preventDefault();

            fetch('/api/logout', {
                method: 'POST',
            })
                .then((response) => {
                    location.reload();
                })
                .catch((error) => {
                    location.reload();
                });
        });
    }

    const loginForm = document.querySelector('.login-form');
    const loginBtn = document.querySelector('#login-btn');
    const errorMsg = document.querySelector('#error-msg');

    if (loginBtn) {
        loginBtn.addEventListener('click', (e) => {
            e.preventDefault();
            const login = loginForm.querySelector('input[name="login"]').value;
            const passwrd = loginForm.querySelector('input[name="password"]').value;

            const formData = new FormData();
            formData.append('login', login);
            formData.append('password', passwrd);

            fetch('/api/login', {
                method: 'POST',
                body: formData,
            })
                .then((response) => {
                    if (!response.ok) {
                        throw new Error('Произошла ошибка на стороне сервера.');
                    }
                    return response.json();
                })
                .then((data) => {
                    if (data.success === false) {
                        errorMsg.textContent = data.reason;
                    } else {
                        window.location.href = "/";
                    }
                })
                .catch((error) => {
                    errorMsg.textContent = 'Произошла ошибка на стороне сервера.';
                });
        });
    }

    const sendMessageButton = document.getElementById('send-message');

    if (sendMessageButton) {
        sendMessageButton.addEventListener('click', (e) => {
            e.preventDefault();

            const errorMsg = document.getElementById("error-msg")
            const message = document.getElementById('message-textarea').value;
            const topicId = document.location.href.split('/').pop();

            fetch('/api/send-message', {
                method: 'POST',
                body: JSON.stringify({
                    topic_id: Number(topicId),
                    message: message
                }),
            })
                .then((response) => {
                    if (!response.ok) {
                        errorMsg.textContent = 'Произошла ошибка на стороне сервера.';

                        return
                    }

                    return response.json();
                })
                .then((data) => {
                    if (data.success === false) {
                        errorMsg.textContent = data.reason;
                    } else {
                        location.reload();
                    }
                })
                .catch((error) => {
                    errorMsg.textContent = 'Произошла ошибка на стороне сервера.';
                });
        });
    }
}