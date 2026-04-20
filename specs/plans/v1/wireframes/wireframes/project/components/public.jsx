/* global React, MOCK, fmtDate, fmtTime, fmtDay, RsvpBadge, EventTypeBadge, SpecNote, StatusBadge, Breadcrumb, Logo */
const { useState } = React;

// ============================================================
// HOME — public landing
// ============================================================
function HomeScreen({ navigate, role, annotations }) {
  return (
    <>
      <section className="hero">
        <Logo size={128} />
        <h1 style={{ marginTop: 24 }}>Orchestre d'Harmonie de Montrouge</h1>
        <div className="hero__tag">Musique de chambre, concerts et répétitions hebdomadaires depuis 1952</div>
        <p className="hero__body">
          Association loi 1901, l'OHM réunit chaque semaine une quarantaine de musiciens amateurs
          au Beffroi de Montrouge. Concerts de printemps, aubades, cérémonies publiques — rejoignez
          un collectif exigeant et chaleureux.
        </p>
        <div style={{ display: 'flex', gap: 12, justifyContent: 'center' }}>
          {role === 'guest' ? (
            <>
              <a href="#" onClick={e => { e.preventDefault(); navigate('login'); }} className="btn btn--primary btn--lg">Espace musicien</a>
              <a href="mailto:contact@ohm-montrouge.fr" className="btn btn--ghost btn--lg">Nous rejoindre</a>
            </>
          ) : (
            <a href="#" onClick={e => { e.preventDefault(); navigate('events'); }} className="btn btn--primary btn--lg">Voir les événements</a>
          )}
        </div>
      </section>

      <div className="page" style={{ maxWidth: 960 }}>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 24 }}>
          <div>
            <div className="eyebrow" style={{ color: 'var(--ohm-gold-ink)' }}>Formation</div>
            <h3 style={{ marginTop: 6 }}>40 musiciens</h3>
            <p className="muted" style={{ fontSize: 14 }}>Bois, cuivres, percussions. Répertoire du classique à la musique de film.</p>
          </div>
          <div>
            <div className="eyebrow" style={{ color: 'var(--ohm-gold-ink)' }}>Répétitions</div>
            <h3 style={{ marginTop: 6 }}>Mercredi, 20h</h3>
            <p className="muted" style={{ fontSize: 14 }}>Salle du Moulin, 12 rue Amaury Duval. Essais libres toute l'année.</p>
          </div>
          <div>
            <div className="eyebrow" style={{ color: 'var(--ohm-gold-ink)' }}>Direction</div>
            <h3 style={{ marginTop: 6 }}>Jean-Philippe Martel</h3>
            <p className="muted" style={{ fontSize: 14 }}>Chef d'orchestre depuis 2019, diplômé du CRR de Paris.</p>
          </div>
        </div>
      </div>

      <SpecNote show={annotations}>
        <span><strong>07-homepage.md</strong> — page statique, même contenu pour authentifiés et non authentifiés. Pas de redirection automatique.</span>
        <span>Bouton « Se connecter » visible uniquement pour les visiteurs non authentifiés.</span>
      </SpecNote>
    </>
  );
}

// ============================================================
// LOGIN
// ============================================================
function LoginScreen({ navigate, annotations }) {
  return (
    <div className="login-shell" style={{ minHeight: 'calc(100vh - var(--header-h))' }}>
      <div className="login-shell__art">
        <div style={{ display: 'flex', alignItems: 'center', gap: 12, position: 'relative', zIndex: 1 }}>
          <Logo size={56} />
          <div style={{ fontFamily: 'var(--font-serif)', fontSize: 22, lineHeight: 1.1 }}>
            Orchestre d'Harmonie<br/>
            <span style={{ color: 'var(--ohm-gold)' }}>de Montrouge</span>
          </div>
        </div>
        <blockquote style={{ fontFamily: 'var(--font-serif)', fontSize: 28, fontStyle: 'italic', lineHeight: 1.3, margin: 0, position: 'relative', zIndex: 1 }}>
          « La musique est le langage<br/>des émotions. »
          <footer style={{ fontSize: 14, fontStyle: 'normal', color: 'var(--ohm-gold)', marginTop: 16 }}>— Emmanuel Kant</footer>
        </blockquote>
        <div style={{ fontSize: 12, color: 'rgba(255,255,255,.5)', position: 'relative', zIndex: 1 }}>
          Association loi 1901 · Siret 321 654 987 00023
        </div>
      </div>

      <div className="login-shell__form">
        <Breadcrumb items={[{ label: 'Accueil', route: 'home' }, { label: 'Connexion' }]} navigate={navigate} />
        <h1 style={{ fontSize: 36, marginBottom: 8 }}>Espace musicien</h1>
        <p className="muted" style={{ marginBottom: 32 }}>Connectez-vous avec l'adresse e-mail communiquée à votre administrateur.</p>

        <form className="stack" style={{ '--gap': '20px' }} onSubmit={e => { e.preventDefault(); navigate('events'); }}>
          <div className="field">
            <label className="field__label">Adresse e-mail <span className="req">*</span></label>
            <input className="input" type="email" placeholder="marie.dubois@example.com" defaultValue="marie.dubois@ohm-montrouge.fr"/>
          </div>
          <div className="field">
            <label className="field__label" style={{ display: 'flex', justifyContent: 'space-between' }}>
              <span>Mot de passe <span className="req">*</span></span>
              <span style={{ fontSize: 12, color: 'var(--ink-4)' }}>Mot de passe oublié ? Contactez un admin.</span>
            </label>
            <input className="input" type="password" defaultValue="••••••••••••••••"/>
          </div>
          <button type="submit" className="btn btn--primary btn--lg" style={{ width: '100%' }}>Se connecter</button>
          <p className="muted" style={{ fontSize: 13, marginTop: 16 }}>
            Pas encore de compte ? Seul un administrateur peut créer un compte musicien. Contactez{' '}
            <a href="mailto:contact@ohm-montrouge.fr">contact@ohm-montrouge.fr</a>.
          </p>
        </form>
      </div>

      <SpecNote show={annotations}>
        <span>Aucune auto-création de compte : l'admin crée le compte, génère le lien d'invitation, et le transmet manuellement.</span>
        <span>Seuls les comptes <code>active</code> peuvent se connecter.</span>
      </SpecNote>
    </div>
  );
}

// ============================================================
// INVITE — account setup flow
// ============================================================
function InviteScreen({ navigate, annotations, expired = false }) {
  const [consentPhone, setConsentPhone] = useState(false);
  const [consentPrivacy, setConsentPrivacy] = useState(false);

  if (expired) {
    return (
      <div className="page page--narrow" style={{ paddingTop: 48 }}>
        <div className="alert alert--warn" style={{ marginBottom: 24 }}>
          <span className="alert__icon">⚠</span>
          <div>
            <strong>Lien d'invitation expiré ou déjà utilisé</strong>
            <div style={{ marginTop: 4, fontSize: 13 }}>Les liens d'invitation sont valides 7 jours. Demandez à un administrateur de vous générer un nouveau lien.</div>
          </div>
        </div>
        <a href="mailto:contact@ohm-montrouge.fr" className="btn btn--primary">Contacter un administrateur</a>
      </div>
    );
  }

  return (
    <div className="page page--narrow">
      <div className="eyebrow">Bienvenue à l'OHM</div>
      <h1 style={{ marginTop: 6, marginBottom: 8 }}>Activez votre compte</h1>
      <p className="muted" style={{ marginBottom: 32 }}>
        Choisissez votre mot de passe pour finaliser votre inscription. Vous serez ensuite redirigé vers la liste des événements.
      </p>

      <div className="card">
        <div className="card__body">
          <div className="kv kv--stacked">
            <dt>Compte</dt>
            <dd><strong>Hugo Bernard</strong> · hugo.bernard@example.com</dd>
            <dt style={{ marginTop: 16 }}>Instrument principal</dt>
            <dd>Saxophone alto</dd>
          </div>
        </div>
      </div>

      <form className="stack" style={{ '--gap': '20px', marginTop: 24 }} onSubmit={e => { e.preventDefault(); navigate('events'); }}>
        <div className="field">
          <label className="field__label">Mot de passe <span className="req">*</span></label>
          <input className="input" type="password" placeholder="Au moins 22 caractères"/>
          <div className="field__hint">Minimum 22 caractères, mêlant majuscules, minuscules, chiffres et caractères spéciaux.</div>
        </div>
        <div className="field">
          <label className="field__label">Confirmer le mot de passe <span className="req">*</span></label>
          <input className="input" type="password" />
        </div>

        <div style={{ borderTop: '1px solid var(--line)', paddingTop: 16, marginTop: 8 }}>
          <label className="checkbox">
            <input type="checkbox" checked={consentPrivacy} onChange={e => setConsentPrivacy(e.target.checked)} />
            <span>
              Je reconnais avoir pris connaissance de la{' '}
              <a href="#" onClick={e => { e.preventDefault(); navigate('privacy'); }}>politique de confidentialité</a> <span className="req">*</span>
            </span>
          </label>
        </div>

        <div className="card card--quiet">
          <div className="card__body" style={{ padding: 16 }}>
            <label className="checkbox">
              <input type="checkbox" checked={consentPhone} onChange={e => setConsentPhone(e.target.checked)} />
              <span>
                <strong>J'accepte que l'association conserve mon téléphone et mon adresse</strong> pour les communications internes (logistique de concert, convocations). Ces données restent visibles uniquement des administrateurs et peuvent être supprimées sur demande à tout moment.
              </span>
            </label>
          </div>
        </div>

        <button type="submit" className="btn btn--primary btn--lg" disabled={!consentPrivacy} style={{ width: '100%' }}>
          Activer mon compte et me connecter
        </button>
      </form>

      <SpecNote show={annotations}>
        <span>Formulaire unique : mot de passe + acceptation politique + consentement téléphone/adresse.</span>
        <span>Bouton désactivé tant que la politique n'est pas acceptée.</span>
        <span>Si consentement = non : téléphone et adresse restent vides et ne pourront être saisis.</span>
        <span>Sur submit (atomique) : status → active, jeton d'invitation marqué utilisé, connexion automatique, redirection vers /evenements.</span>
      </SpecNote>
    </div>
  );
}

// ============================================================
// PASSWORD RESET
// ============================================================
function PasswordResetScreen({ navigate, annotations }) {
  return (
    <div className="page page--narrow">
      <h1 style={{ marginBottom: 8 }}>Nouveau mot de passe</h1>
      <p className="muted" style={{ marginBottom: 32 }}>Choisissez un nouveau mot de passe pour votre compte.</p>

      <form className="stack" style={{ '--gap': '20px' }} onSubmit={e => { e.preventDefault(); navigate('login'); }}>
        <div className="field">
          <label className="field__label">Nouveau mot de passe <span className="req">*</span></label>
          <input className="input" type="password" />
          <div className="field__hint">Minimum 22 caractères, mêlant majuscules, minuscules, chiffres et caractères spéciaux.</div>
        </div>
        <div className="field">
          <label className="field__label">Confirmer <span className="req">*</span></label>
          <input className="input" type="password" />
        </div>
        <button type="submit" className="btn btn--primary btn--lg">Mettre à jour</button>
      </form>

      <SpecNote show={annotations}>
        <span>Le lien de réinitialisation est généré par un admin depuis la fiche du musicien. Expire après 7 jours.</span>
      </SpecNote>
    </div>
  );
}

// ============================================================
// PRIVACY NOTICE
// ============================================================
function PrivacyScreen({ annotations }) {
  return (
    <div className="page page--narrow">
      <div className="eyebrow">RGPD — Article 13</div>
      <h1 style={{ marginTop: 6 }}>Politique de confidentialité</h1>
      <p className="muted">Dernière mise à jour : 1er septembre 2025</p>

      <div className="stack" style={{ '--gap': '28px', marginTop: 32 }}>
        <section>
          <h3>1. Responsable du traitement</h3>
          <p>L'Orchestre d'Harmonie de Montrouge (OHM), association loi 1901 sise au 12 rue Amaury Duval, 92120 Montrouge, traite les données personnelles des musiciens membres.</p>
        </section>
        <section>
          <h3>2. Données collectées</h3>
          <ul>
            <li><strong>Obligatoires</strong> : prénom, nom, e-mail, instrument principal.</li>
            <li><strong>Facultatives</strong> : date de naissance, téléphone, adresse postale (sous consentement explicite).</li>
            <li><strong>Mineurs de moins de 15 ans</strong> : un URI de consentement parental est requis (Art. 8).</li>
          </ul>
        </section>
        <section>
          <h3>3. Finalités et base légale</h3>
          <p>Gestion de l'association (intérêt légitime), communications logistiques (consentement), suivi des cotisations (exécution contractuelle).</p>
        </section>
        <section>
          <h3>4. Durée de conservation</h3>
          <p>Les données sont conservées pendant la durée de l'adhésion, puis 5 ans à compter de la fin de la dernière saison cotisée, conformément aux obligations comptables de l'association.</p>
        </section>
        <section>
          <h3>5. Vos droits</h3>
          <p>Vous pouvez exercer vos droits d'accès, de rectification, d'effacement (anonymisation), de limitation et de portabilité en contactant un administrateur à <a href="mailto:rgpd@ohm-montrouge.fr">rgpd@ohm-montrouge.fr</a>.</p>
        </section>
      </div>

      <SpecNote show={annotations}>
        <span>Page statique, bundlée avec l'application. Pas d'admin UI en V1.</span>
        <span>Accessible publiquement (via lien d'invitation sans être connecté).</span>
      </SpecNote>
    </div>
  );
}

Object.assign(window, { HomeScreen, LoginScreen, InviteScreen, PasswordResetScreen, PrivacyScreen });
