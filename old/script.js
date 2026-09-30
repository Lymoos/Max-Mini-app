const views = document.querySelectorAll('.view');
const navButtons = document.querySelectorAll('[data-target]');

function showView(name) {
  views.forEach(view => {
    view.hidden = view.dataset.view !== name;
  });
  document.querySelectorAll('.nav-item').forEach(item => {
    item.classList.toggle('nav-item-active', item.dataset.target === name);
  });
}

navButtons.forEach(btn => {
  btn.addEventListener('click', () => showView(btn.dataset.target));
});
