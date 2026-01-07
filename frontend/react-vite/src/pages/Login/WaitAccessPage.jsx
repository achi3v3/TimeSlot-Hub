import LeftSidebar from "../../components/LeftSidebar/LeftSidebar";
import WaitAccess from "../../components/waitAccess/waitAccess";
import { useState } from "react";

export const WaitAccessPage = () => {
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);

  const toggleMobileMenu = () => {
    setIsMobileMenuOpen(!isMobileMenuOpen);
  };
  return (
    <div className="login-root">
      {/* Мобильная кнопка меню */}
      <button 
        className="mobile-menu-toggle"
        onClick={toggleMobileMenu}
        aria-label="Открыть меню"
      >
        <span className={`hamburger ${isMobileMenuOpen ? 'active' : ''}`}>
          <span></span>
          <span></span>
          <span></span>
        </span>
      </button>

      {/* Мобильный оверлей */}
      {isMobileMenuOpen && (
        <div 
          className="mobile-overlay"
          onClick={toggleMobileMenu}
        ></div>
      )}

      {/* Левый сайдбар навигации */}
      <LeftSidebar 
        isMobileMenuOpen={isMobileMenuOpen}
        toggleMobileMenu={toggleMobileMenu}
      />
      
      <div className="login-content-wrapper">
        <WaitAccess />
      </div>
    </div>
  );
};
