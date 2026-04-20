/* global React, ReactDOM, AppBar, Footer, HomeScreen, LoginScreen, InviteScreen, PasswordResetScreen, PrivacyScreen,
          EventListScreen, EventDetailScreen, ProfileScreen,
          AdminMusiciansScreen, AdminMusicianDetailScreen, AdminMusicianNewScreen,
          AdminEventsScreen, AdminEventEditScreen, AdminSeasonsScreen, AdminRetentionScreen */
const { useState, useEffect } = React;

const TWEAK_DEFAULTS = /*EDITMODE-BEGIN*/{
  "role": "admin",
  "accountState": "active",
  "dataState": "populated",
  "eventType": "concert",
  "annotations": true,
  "density": "comfortable",
  "viewMode": "desktop"
}/*EDITMODE-END*/;

function App() {
  const [route, setRoute] = useState(() => localStorage.getItem('ohm-route') || 'home');
  const [params, setParams] = useState({});
  const [tweaks, setTweaks] = useState(TWEAK_DEFAULTS);
  const [editMode, setEditMode] = useState(false);

  // Edit mode protocol
  useEffect(() => {
    const onMsg = (e) => {
      if (!e.data || !e.data.type) return;
      if (e.data.type === '__activate_edit_mode') setEditMode(true);
      if (e.data.type === '__deactivate_edit_mode') setEditMode(false);
    };
    window.addEventListener('message', onMsg);
    window.parent.postMessage({ type: '__edit_mode_available' }, '*');
    return () => window.removeEventListener('message', onMsg);
  }, []);

  useEffect(() => { localStorage.setItem('ohm-route', route); }, [route]);
  useEffect(() => { document.documentElement.dataset.density = tweaks.density; }, [tweaks.density]);

  const navigate = (r, p = {}) => {
    setRoute(r);
    setParams(p);
    if (p.eventType) setTweaks(t => ({ ...t, eventType: p.eventType }));
    window.scrollTo(0, 0);
  };

  const setTweak = (key, val) => {
    setTweaks(t => {
      const next = { ...t, [key]: val };
      window.parent.postMessage({ type: '__edit_mode_set_keys', edits: { [key]: val } }, '*');
      return next;
    });
  };

  const { role, accountState, dataState, eventType, annotations, viewMode } = tweaks;

  // Route the screen
  let screen;
  const navProps = { navigate, annotations };
  switch (route) {
    case 'home':        screen = <HomeScreen {...navProps} role={role}/>; break;
    case 'login':       screen = <LoginScreen {...navProps}/>; break;
    case 'invite':      screen = <InviteScreen {...navProps} expired={false}/>; break;
    case 'invite-expired': screen = <InviteScreen {...navProps} expired={true}/>; break;
    case 'password-reset': screen = <PasswordResetScreen {...navProps}/>; break;
    case 'privacy':     screen = <PrivacyScreen {...navProps}/>; break;
    case 'events':      screen = <EventListScreen {...navProps} state={dataState}/>; break;
    case 'event-detail': screen = <EventDetailScreen {...navProps} role={role} eventType={params.eventType || eventType}/>; break;
    case 'profile':     screen = <ProfileScreen {...navProps}/>; break;
    case 'admin-musicians': screen = <AdminMusiciansScreen {...navProps} state={dataState}/>; break;
    case 'admin-musician-detail': screen = <AdminMusicianDetailScreen {...navProps} accountState={accountState}/>; break;
    case 'admin-musician-new': screen = <AdminMusicianNewScreen {...navProps}/>; break;
    case 'admin-events': screen = <AdminEventsScreen {...navProps}/>; break;
    case 'admin-event-edit': screen = <AdminEventEditScreen {...navProps} eventType={params.eventType || eventType}/>; break;
    case 'admin-seasons': screen = <AdminSeasonsScreen {...navProps}/>; break;
    case 'admin-retention': screen = <AdminRetentionScreen {...navProps}/>; break;
    default: screen = <HomeScreen {...navProps} role={role}/>;
  }

  const showChrome = !['login','invite','invite-expired','password-reset'].includes(route);
  const mobile = viewMode === 'mobile';

  const content = (
    <div className={`viewport ${mobile ? 'viewport--mobile' : ''}`}>
      {showChrome && <AppBar role={role} route={route} navigate={navigate} />}
      <main style={{ flex: 1 }}>{screen}</main>
      {showChrome && <Footer onPrivacy={() => navigate('privacy')} />}
    </div>
  );

  return (
    <>
      {content}
      {editMode && <TweaksPanel tweaks={tweaks} setTweak={setTweak} route={route} navigate={navigate}/>}
    </>
  );
}

function TweaksPanel({ tweaks, setTweak, route, navigate }) {
  const Opt = ({ k, v, label }) => (
    <button className={`tweak__opt ${tweaks[k]===v ? 'tweak__opt--active' : ''}`} onClick={() => setTweak(k, v)}>{label}</button>
  );
  return (
    <div className="tweaks">
      <div className="tweaks__header"><h4>Tweaks</h4><span style={{ fontSize: 11, color: 'var(--ink-4)' }}>OHM Wireframes</span></div>
      <div className="tweaks__body">
        <div className="tweak">
          <div className="tweak__label">Rôle / authentification</div>
          <div className="tweak__options">
            <Opt k="role" v="guest" label="Invité"/>
            <Opt k="role" v="musician" label="Musicien"/>
            <Opt k="role" v="admin" label="Admin"/>
          </div>
        </div>
        <div className="tweak">
          <div className="tweak__label">État du compte (fiche musicien)</div>
          <div className="tweak__options">
            <Opt k="accountState" v="pending" label="En attente"/>
            <Opt k="accountState" v="active" label="Actif"/>
            <Opt k="accountState" v="anonymized" label="Anonymisé"/>
          </div>
        </div>
        <div className="tweak">
          <div className="tweak__label">Données</div>
          <div className="tweak__options">
            <Opt k="dataState" v="populated" label="Avec données"/>
            <Opt k="dataState" v="empty" label="Vide"/>
          </div>
        </div>
        <div className="tweak">
          <div className="tweak__label">Type d'événement (détail)</div>
          <div className="tweak__options">
            <Opt k="eventType" v="concert" label="Concert"/>
            <Opt k="eventType" v="rehearsal" label="Répétition"/>
            <Opt k="eventType" v="other" label="Autre"/>
          </div>
        </div>
        <div className="tweak">
          <div className="tweak__label">Annotations spec</div>
          <div className="tweak__options">
            <Opt k="annotations" v={true} label="Visible"/>
            <Opt k="annotations" v={false} label="Masqué"/>
          </div>
        </div>
        <div className="tweak">
          <div className="tweak__label">Densité</div>
          <div className="tweak__options">
            <Opt k="density" v="comfortable" label="Confortable"/>
            <Opt k="density" v="compact" label="Compacte"/>
          </div>
        </div>
        <div className="tweak">
          <div className="tweak__label">Aperçu</div>
          <div className="tweak__options">
            <Opt k="viewMode" v="desktop" label="Desktop"/>
            <Opt k="viewMode" v="mobile" label="Mobile"/>
          </div>
        </div>
        <div className="tweak">
          <div className="tweak__label">Aller à l'écran</div>
          <select className="select" value={route} onChange={e => navigate(e.target.value)}>
            <optgroup label="Public">
              <option value="home">Accueil</option>
              <option value="login">Connexion</option>
              <option value="invite">Invitation — activation</option>
              <option value="invite-expired">Invitation — expirée</option>
              <option value="password-reset">Réinit. mot de passe</option>
              <option value="privacy">Politique confidentialité</option>
            </optgroup>
            <optgroup label="Musicien">
              <option value="events">Liste événements</option>
              <option value="event-detail">Détail événement</option>
              <option value="profile">Mon profil</option>
            </optgroup>
            <optgroup label="Admin">
              <option value="admin-musicians">Musiciens</option>
              <option value="admin-musician-detail">Fiche musicien</option>
              <option value="admin-musician-new">Nouveau musicien</option>
              <option value="admin-events">Événements admin</option>
              <option value="admin-event-edit">Édition événement</option>
              <option value="admin-seasons">Saisons</option>
              <option value="admin-retention">Rétention RGPD</option>
            </optgroup>
          </select>
        </div>
      </div>
    </div>
  );
}

ReactDOM.createRoot(document.getElementById('root')).render(<App/>);
