export default function SidebarToggle({ onToggle }) {
  return (
    <button className="sidebar-toggle" onClick={onToggle} aria-label="Открыть меню">
      <span className="hamburger-line" />
      <span className="hamburger-line" />
      <span className="hamburger-line" />
    </button>
  );
}


