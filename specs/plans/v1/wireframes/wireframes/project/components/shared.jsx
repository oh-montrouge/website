/* global React */
const { useState } = React;

// ============================================================
// LOGO — inline SVG badge of the Montrouge bell tower, stylized
// ============================================================
function Logo({ size = 36 }) {
  return (
    <svg width={size} height={size} viewBox="0 0 64 64" fill="none" style={{ display: 'block' }} aria-hidden="true">
      <circle cx="32" cy="32" r="31" fill="#fbfaf7" stroke="#8b1e3f" strokeWidth="1.2"/>
      {/* city skyline */}
      <path d="M10 46 L10 40 L14 40 L14 36 L18 36 L18 40 L22 40 L22 42 L26 42 L26 32 L28 32 L28 28 L30 28 L30 24 L32 22 L34 24 L34 28 L36 28 L36 32 L38 32 L38 42 L42 42 L42 36 L46 36 L46 40 L50 40 L50 42 L54 42 L54 46 Z"
            fill="#bca15d"/>
      {/* clock */}
      <circle cx="32" cy="32" r="4" fill="#fbfaf7" stroke="#bca15d" strokeWidth="1"/>
      <line x1="32" y1="32" x2="32" y2="30" stroke="#8b1e3f" strokeWidth="1" strokeLinecap="round"/>
      <line x1="32" y1="32" x2="33.5" y2="32" stroke="#8b1e3f" strokeWidth="1" strokeLinecap="round"/>
      {/* OHM mark */}
      <text x="32" y="55" fontFamily="Cormorant Garamond, Georgia, serif" fontSize="7" fontWeight="600" fill="#8b1e3f" textAnchor="middle" letterSpacing="1.5">OHM</text>
    </svg>
  );
}

// ============================================================
// APPBAR — with auth-state-aware nav
// ============================================================
function AppBar({ role, route, navigate, hasSheetMusic = true }) {
  const isAuth = role !== 'guest';
  const isAdmin = role === 'admin';

  return (
    <header className="appbar">
      <button className="appbar__menu" aria-label="Menu" onClick={e => e.preventDefault()}>☰</button>
      <div className="appbar__brand" onClick={() => navigate('home')} style={{ cursor: 'pointer' }}>
        <Logo size={40} />
        <div className="appbar__brand-text">
          <b>OHM</b>
          <span className="appbar__brand-tagline">Orchestre d'Harmonie<br/>de Montrouge</span>
        </div>
      </div>
      {isAuth ? (
        <>
          <nav className="appbar__nav">
            <a className={`appbar__nav-item ${route === 'events' ? 'appbar__nav-item--active' : ''}`} onClick={e => { e.preventDefault(); navigate('events'); }} href="#">Événements</a>
            {hasSheetMusic && (
              <a className="appbar__nav-item" href="https://drive.google.com" target="_blank" rel="noreferrer">Partitions ↗</a>
            )}
            <a className={`appbar__nav-item ${route === 'profile' ? 'appbar__nav-item--active' : ''}`} onClick={e => { e.preventDefault(); navigate('profile'); }} href="#">Mon profil</a>
            {isAdmin && (
              <>
                <span style={{ width: 1, height: 22, background: 'var(--line)', margin: '0 6px' }}></span>
                <a className={`appbar__nav-item appbar__nav-item--admin ${route.startsWith('admin-musicians') ? 'appbar__nav-item--active' : ''}`} onClick={e => { e.preventDefault(); navigate('admin-musicians'); }} href="#">Musiciens</a>
                <a className={`appbar__nav-item appbar__nav-item--admin ${route.startsWith('admin-events') ? 'appbar__nav-item--active' : ''}`} onClick={e => { e.preventDefault(); navigate('admin-events'); }} href="#">Événements (admin)</a>
                <a className={`appbar__nav-item appbar__nav-item--admin ${route === 'admin-seasons' ? 'appbar__nav-item--active' : ''}`} onClick={e => { e.preventDefault(); navigate('admin-seasons'); }} href="#">Saisons</a>
                <a className={`appbar__nav-item appbar__nav-item--admin ${route === 'admin-retention' ? 'appbar__nav-item--active' : ''}`} onClick={e => { e.preventDefault(); navigate('admin-retention'); }} href="#">Rétention</a>
              </>
            )}
          </nav>
          <div className="appbar__user">
            <span className="avatar">MD</span>
            <div className="appbar__user-name" style={{ lineHeight: 1.2 }}>
              <div style={{ fontSize: 13, fontWeight: 500 }}>
                Marie Dubois
                {isAdmin && <span className="badge badge--admin" style={{ marginLeft: 6, fontSize: 10, padding: '1px 6px' }}>Admin</span>}
              </div>
              <a href="#" onClick={e => { e.preventDefault(); navigate('home'); }} style={{ fontSize: 11, color: 'var(--ink-3)' }}>Déconnexion</a>
            </div>
          </div>
        </>
      ) : (
        <>
          <div className="appbar__nav"></div>
          <a href="#" onClick={e => { e.preventDefault(); navigate('login'); }} className="btn btn--primary btn--sm">Se connecter</a>
        </>
      )}
    </header>
  );
}

// ============================================================
// FOOTER
// ============================================================
function Footer({ onPrivacy }) {
  return (
    <footer className="footer">
      <div>
        © 2025 · Orchestre d'Harmonie de Montrouge · Association loi 1901
      </div>
      <div style={{ display: 'flex', gap: 16 }}>
        <a href="#" onClick={e => { e.preventDefault(); onPrivacy && onPrivacy(); }}>Politique de confidentialité</a>
        <span style={{ color: 'var(--ink-5)' }}>·</span>
        <a href="mailto:contact@ohm-montrouge.fr">contact@ohm-montrouge.fr</a>
      </div>
    </footer>
  );
}

// ============================================================
// BADGES for account / RSVP / event status
// ============================================================
function StatusBadge({ status }) {
  const map = {
    pending:    { cls: 'badge--warn',    label: 'En attente' },
    active:     { cls: 'badge--ok',      label: 'Actif' },
    anonymized: { cls: 'badge--neutral', label: 'Anonymisé' },
  };
  const s = map[status] || map.pending;
  return <span className={`badge badge--dot ${s.cls}`}>{s.label}</span>;
}
function RsvpBadge({ state }) {
  if (!state || state === 'unanswered') return <span className="badge badge--neutral">Sans réponse</span>;
  const map = {
    yes:   { cls: 'badge--ok',      label: '✓ Présent' },
    no:    { cls: 'badge--neutral', label: '✗ Absent' },
    maybe: { cls: 'badge--warn',    label: '? Peut‑être' },
  };
  return <span className={`badge ${map[state].cls}`}>{map[state].label}</span>;
}
function EventTypeBadge({ type }) {
  const map = {
    concert:   { cls: 'badge--gold',    label: 'Concert' },
    rehearsal: { cls: 'badge--info',    label: 'Répétition' },
    other:     { cls: 'badge--neutral', label: 'Autre' },
  };
  const s = map[type] || map.other;
  return <span className={`badge ${s.cls}`}>{s.label}</span>;
}

// ============================================================
// SPEC NOTE — subtle margin annotations
// ============================================================
function SpecNote({ children, show }) {
  if (!show) return null;
  return <div className="annotations"><h4>Règle métier</h4><ul>{React.Children.map(children, c => <li>{c}</li>)}</ul></div>;
}

// ============================================================
// Breadcrumb
// ============================================================
function Breadcrumb({ items, navigate }) {
  return (
    <div className="breadcrumb">
      {items.map((it, i) => (
        <React.Fragment key={i}>
          {i > 0 && <span className="breadcrumb__sep">›</span>}
          {it.route ? (
            <a href="#" onClick={e => { e.preventDefault(); navigate(it.route); }}>{it.label}</a>
          ) : (
            <span>{it.label}</span>
          )}
        </React.Fragment>
      ))}
    </div>
  );
}

// ============================================================
// Export
// ============================================================
Object.assign(window, { Logo, AppBar, Footer, StatusBadge, RsvpBadge, EventTypeBadge, SpecNote, Breadcrumb });
