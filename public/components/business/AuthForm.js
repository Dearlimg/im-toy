// 业务组件：登录/注册表单
// 依赖：services/api、store/auth、utils/validator、Toast

import { register, login } from '../../services/api.js';
import { saveSession } from '../../store/auth.js';
import { validUsername, validNickname, validPassword } from '../../utils/validator.js';
import { Toast } from '../common/Toast.js';

export function initAuthForm({ onSuccess }) {
  const msg = document.getElementById('authMsg');
  const loginForm = document.getElementById('loginForm');
  const registerForm = document.getElementById('registerForm');
  const tabLogin = document.getElementById('tabLogin');
  const tabReg = document.getElementById('tabReg');
  const hint = document.getElementById('apiHint');

  const cfg = (typeof window.IM_CONFIG !== 'undefined' && window.IM_CONFIG) || {};
  if (hint) hint.textContent = cfg.API_ORIGIN ? 'API: ' + cfg.API_ORIGIN : '';

  function showTab(mode) {
    loginForm.classList.toggle('hidden', mode !== 'login');
    registerForm.classList.toggle('hidden', mode !== 'register');
    tabLogin.classList.toggle('active', mode === 'login');
    tabReg.classList.toggle('active', mode === 'register');
  }

  function setMsg(text) {
    msg.textContent = text || '';
  }

  async function doLogin() {
    const username = document.getElementById('loginUsername').value.trim();
    const password = document.getElementById('loginPassword').value;
    if (!validUsername(username)) return setMsg('用户名需为 3-32 位字母数字或下划线');
    if (!validPassword(password)) return setMsg('密码至少 6 位');
    setMsg('登录中…');
    try {
      const data = await login(username, password);
      saveSession(data.token, data.user || { username });
      Toast('登录成功');
      onSuccess(data);
    } catch (e) {
      setMsg(e.message);
    }
  }

  async function doRegister() {
    const username = document.getElementById('regUsername').value.trim();
    const nickname = document.getElementById('regNickname').value.trim();
    const password = document.getElementById('regPassword').value;
    if (!validUsername(username)) return setMsg('用户名需为 3-32 位字母数字或下划线');
    if (!validNickname(nickname)) return setMsg('昵称不能为空且不超过 32 字');
    if (!validPassword(password)) return setMsg('密码至少 6 位');
    setMsg('注册中…');
    try {
      await register(username, nickname, password);
      setMsg('注册成功，请登录');
      showTab('login');
      document.getElementById('loginUsername').value = username;
      document.getElementById('loginPassword').value = '';
    } catch (e) {
      setMsg(e.message);
    }
  }

  tabLogin.addEventListener('click', () => { setMsg(''); showTab('login'); });
  tabReg.addEventListener('click', () => { setMsg(''); showTab('register'); });
  document.getElementById('loginBtn')?.addEventListener('click', doLogin);
  document.getElementById('regBtn')?.addEventListener('click', doRegister);

  return { showTab, doLogin, doRegister };
}
