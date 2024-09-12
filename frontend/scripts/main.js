
window.onload = function() {
    const remove_avatar_button = document.getElementById('remove-avatar-button');

    if (remove_avatar_button) {
        remove_avatar_button.addEventListener('click', (e) => {
            e.preventDefault();

            const newValues = {
                avatarRemove: true,
                username: form.querySelector('input[placeholder="Юзернейм"]').value,
                description: form.querySelector('textarea[placeholder="Текст описания"]').value,
                signText: form.querySelector('textarea[placeholder="Текст подписи"]').value,
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
                description: form.querySelector('textarea[placeholder="Текст описания"]').value,
                avatar: null,
                signText: form.querySelector('textarea[placeholder="Текст подписи"]').value,
            };

            saveButton.addEventListener('click', (e) => {
                e.preventDefault();

                const newValues = {
                    username: form.querySelector('input[placeholder="Юзернейм"]').value,
                    description: form.querySelector('textarea[placeholder="Текст описания"]').value,
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
    const saveButton2 = document.getElementById('create-topic-btn');
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

    const menu_button = document.getElementById('menu-button');

    const menu = document.getElementById("menu");
    let showed = false;

    if (menu_button) {
        menu_button.addEventListener('click', (e) => {
            e.preventDefault();

            if (menu.classList.contains("menu-container-open")) {
                menu_button.classList.remove("account-button-opened")
                menu.classList.remove("menu-container-open")
                showed = false;
            } else {
                menu_button.classList.add("account-button-opened")
                menu.classList.add("menu-container-open")
                showed = true;
            }
        })

        let left = getCoords(menu_button).left + menu_button.offsetWidth - 150;
        let existsTimeout = -1

        window.onresize = function() {
            left = getCoords(menu_button).left + menu_button.offsetWidth - 150;
            menu.style.margin = 'auto ' + left + 'px';
            menu.style.transition = '0s';

            clearTimeout(existsTimeout)

            existsTimeout = setTimeout(function() {
                menu.style.transition = '0.12s';
            }, 100)
        }

        menu.style.margin = 'auto ' + left + 'px';
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
    const registerBtn = document.querySelector('#register-btn');
    const errorMsg = document.querySelector('#error-msg');

    if (registerBtn) {
        let btn_activated = false;

        registerBtn.addEventListener('click', (e) => {
            e.preventDefault();
            if (btn_activated) { return }
            const login = loginForm.querySelector('input[name="login"]').value;
            const password = loginForm.querySelector('input[name="password"]').value;
            const username = loginForm.querySelector('input[name="username"]').value;
            const email = loginForm.querySelector('input[name="email"]').value;

            const formData = new FormData();
            formData.append('login', login);
            formData.append('password', password);
            formData.append('username', username);
            formData.append('email', email);

            btn_activated = true
            fetch('/api/register', {
                method: 'POST',
                body: formData,
            })
                .then((response) => {
                    if (!response.ok) {
                        btn_activated = false
                        throw new Error('Произошла ошибка на стороне сервера.');
                    }

                    return response.json();
                })
                .then((data) => {
                    if (data.success === false) {
                        btn_activated = false
                        errorMsg.textContent = data.reason;
                    } else {
                        errorMsg.textContent = "Вы успешно зарегистрировались! Сейчас мы перенаправим вам на страницу с авторизацией..."
                        setInterval(function() {
                            window.location.href = "/login"
                        }, 1500)
                    }
                })
                .catch((error) => {
                    btn_activated = false
                    errorMsg.textContent = 'Произошла ошибка на стороне сервера.';
                });
        });
    }

    const loginBtn = document.querySelector('#login-btn');

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
            let topicId = document.location.href.split('/');

            fetch('/api/send-message', {
                method: 'POST',
                body: JSON.stringify({
                    topic_id: Number(topicId[topicId.length - 2]),
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