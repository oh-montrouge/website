/* global React, MOCK, fmtDate, fmtTime, StatusBadge, EventTypeBadge, RsvpBadge, SpecNote, Breadcrumb */
const { useState } = React;

// ============================================================
// MUSICIANS LIST (admin)
// ============================================================
function AdminMusiciansScreen({ navigate, annotations, state = 'populated' }) {
  const [filter, setFilter] = useState('all');
  const [search, setSearch] = useState('');

  const filtered = MOCK.musicians.filter(m => {
    if (filter !== 'all' && m.status !== filter) return false;
    if (search && !`${m.first} ${m.last}`.toLowerCase().includes(search.toLowerCase())) return false;
    return true;
  });

  return (
    <div className="page page--wide">
      <div className="page__header">
        <div className="page__title">
          <h1>Musiciens</h1>
          <p>Gestion des comptes membres : création, invitation, anonymisation.</p>
        </div>
        <div className="page__actions">
          <button className="btn btn--ghost" onClick={() => navigate('admin-retention')}>Revue de rétention (3)</button>
          <button className="btn btn--primary" onClick={() => navigate('admin-musician-new')}>+ Nouveau musicien</button>
        </div>
      </div>

      <div className="stat-grid" style={{ marginBottom: 24 }}>
        <div className="stat"><div className="stat__label">Actifs</div><div className="stat__value">37</div></div>
        <div className="stat"><div className="stat__label">En attente d'invitation</div><div className="stat__value">2</div></div>
        <div className="stat"><div className="stat__label">Admins</div><div className="stat__value">3</div><div className="stat__delta" style={{ color: 'var(--ok)' }}>Protection min. 1</div></div>
        <div className="stat"><div className="stat__label">Anonymisés</div><div className="stat__value">12</div></div>
      </div>

      <div className="row" style={{ gap: 12, marginBottom: 16 }}>
        <input className="input" placeholder="Rechercher un musicien..." value={search} onChange={e => setSearch(e.target.value)} style={{ maxWidth: 320 }}/>
        <div className="tabs" style={{ margin: 0, border: 'none' }}>
          {['all','active','pending','anonymized'].map(f => (
            <button key={f} className={`tab ${filter===f?'tab--active':''}`} onClick={() => setFilter(f)}>
              {f==='all'?'Tous':f==='active'?'Actifs':f==='pending'?'En attente':'Anonymisés'}
            </button>
          ))}
        </div>
      </div>

      {state === 'empty' ? (
        <div className="empty">
          <div className="empty__icon">♫</div>
          <h3>Aucun musicien</h3>
          <p>Commencez par créer le premier compte musicien. Un lien d'invitation sera généré à partager avec la personne.</p>
          <button className="btn btn--primary" style={{ marginTop: 16 }} onClick={() => navigate('admin-musician-new')}>Créer un musicien</button>
        </div>
      ) : (
        <div className="card" style={{ overflow: 'hidden' }}>
          <table className="table">
            <thead>
              <tr>
                <th>Nom</th>
                <th>Instrument</th>
                <th>E-mail</th>
                <th>Première inscription</th>
                <th>Statut</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {filtered.map(m => (
                <tr key={m.id} style={{ cursor: 'pointer' }} onClick={() => navigate('admin-musician-detail', { id: m.id })}>
                  <td data-label="Musicien">
                    <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                      <span className={`avatar ${m.status === 'anonymized' ? 'avatar--muted' : ''}`}>
                        {m.status === 'anonymized' ? '?' : `${m.first[0]}${m.last[0]}`}
                      </span>
                      <div>
                        <div style={{ fontWeight: 500 }}>
                          {m.status === 'anonymized' ? <span className="muted">(anonymisé)</span> : `${m.first} ${m.last}`}
                          {m.admin && <span className="badge badge--admin" style={{ marginLeft: 8, fontSize: 10 }}>Admin</span>}
                        </div>
                      </div>
                    </div>
                  </td>
                  <td data-label="Instrument">{m.instrument}</td>
                  <td data-label="E-mail" className="muted mono" style={{ fontSize: 12 }}>{m.email || '—'}</td>
                  <td data-label="Inscription" className="muted">{m.inscription || '—'}</td>
                  <td data-label="Statut"><StatusBadge status={m.status} /></td>
                  <td className="table__actions"><span style={{ color: 'var(--ink-4)' }}>›</span></td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      <SpecNote show={annotations}>
        <span>Invariant I3 : au moins un compte <code>active</code> doit porter le rôle admin à tout moment.</span>
        <span>Les comptes anonymisés restent listés (pour l'historique des cotisations) mais ne peuvent pas se connecter.</span>
      </SpecNote>
    </div>
  );
}

// ============================================================
// MUSICIAN DETAIL (admin)
// ============================================================
function AdminMusicianDetailScreen({ navigate, annotations, accountState = 'active' }) {
  const [showInvite, setShowInvite] = useState(false);
  const [showAnonymize, setShowAnonymize] = useState(false);
  const [showBlocked, setShowBlocked] = useState(false);
  const [restricted, setRestricted] = useState(false);

  const isPending = accountState === 'pending';
  const isAnonymized = accountState === 'anonymized';

  return (
    <div className="page page--wide">
      <Breadcrumb items={[{ label: 'Musiciens', route: 'admin-musicians' }, { label: isAnonymized ? '(anonymisé)' : 'Marie Dubois' }]} navigate={navigate}/>
      <div className="page__header">
        <div className="page__title">
          <div style={{ display: 'flex', alignItems: 'center', gap: 16, marginBottom: 8 }}>
            <span className="avatar avatar--lg">{isAnonymized ? '?' : 'MD'}</span>
            <div>
              <div style={{ display: 'flex', gap: 8, alignItems: 'center', marginBottom: 4 }}>
                <StatusBadge status={accountState} />
                {!isAnonymized && <span className="badge badge--admin">Admin</span>}
                {restricted && <span className="badge badge--warn">Traitement restreint (Art. 18)</span>}
              </div>
              <h1 style={{ fontSize: 28 }}>{isAnonymized ? '— Compte anonymisé —' : 'Marie Dubois'}</h1>
              <div className="muted">Clarinette {!isAnonymized && '· 1ère inscription : septembre 2018'}</div>
            </div>
          </div>
        </div>
        <div className="page__actions">
          {!isAnonymized && <button className="btn btn--ghost" onClick={() => navigate('admin-musician-edit')}>Modifier</button>}
          {isPending && <button className="btn btn--danger" onClick={() => setShowBlocked(true)}>Supprimer</button>}
          {accountState === 'active' && <button className="btn btn--danger" onClick={() => setShowAnonymize(true)}>Anonymiser (RGPD)</button>}
        </div>
      </div>

      {isPending && (
        <div className="alert alert--warn" style={{ marginBottom: 24 }}>
          <span className="alert__icon">⏳</span>
          <div style={{ flex: 1 }}>
            <strong>Compte en attente d'activation</strong>
            <div style={{ fontSize: 13, marginTop: 4 }}>Le musicien n'a pas encore complété son inscription. Un lien d'invitation a été généré il y a 2 jours — expire dans 5 jours.</div>
          </div>
          <button className="btn btn--sm" onClick={() => setShowInvite(true)}>Voir / régénérer le lien</button>
        </div>
      )}

      {isAnonymized && (
        <div className="alert alert--info" style={{ marginBottom: 24 }}>
          <span className="alert__icon">🛡</span>
          <div>
            <strong>Compte anonymisé le 12 janvier 2026</strong>
            <div style={{ fontSize: 13, marginTop: 4 }}>Nom, e-mail, téléphone, adresse et mot de passe ont été effacés. L'instrument principal et l'historique des cotisations (pseudonymisé) sont conservés. Les RSVP ont été supprimés.</div>
          </div>
        </div>
      )}

      <div className="split">
        <div className="stack" style={{ '--gap': '24px' }}>
          <div className="card">
            <div className="card__header"><h4>Informations</h4></div>
            <div className="card__body">
              <dl className="kv">
                <dt>Prénom, nom</dt><dd>{isAnonymized ? <span className="muted">—</span> : 'Marie Dubois'}</dd>
                <dt>E-mail</dt><dd>{isAnonymized ? <span className="muted">—</span> : 'marie.dubois@ohm-montrouge.fr'}</dd>
                <dt>Instrument principal</dt><dd>Clarinette</dd>
                <dt>Date de naissance</dt><dd>{isAnonymized ? <span className="muted">—</span> : '14 mars 1987'}</dd>
                {!isAnonymized && <><dt>Consentement téléphone/adresse</dt><dd><span className="badge badge--ok">✓ Donné</span></dd></>}
                {!isAnonymized && <><dt>Téléphone</dt><dd>06 12 34 56 78</dd></>}
                {!isAnonymized && <><dt>Adresse</dt><dd>42 rue de la Paix, 92120 Montrouge</dd></>}
                {!isAnonymized && <><dt>Traitement restreint</dt><dd>
                  {restricted ? (
                    <span className="badge" style={{ background: 'var(--warn-bg)', color: 'var(--warn)', borderColor: '#e6d09a' }}>
                      ⊘ Article 18 RGPD appliqué
                    </span>
                  ) : (
                    <span className="muted" style={{ fontSize: 13 }}>Aucune restriction</span>
                  )}
                </dd></>}
                {isAnonymized && <><dt>Jeton d'anonymisation</dt><dd className="mono" style={{ fontSize: 12 }}>anon_7f4b8c2d9e1a</dd></>}
              </dl>
            </div>
          </div>

          {/* Fee payments */}
          <div className="card">
            <div className="card__header">
              <h4>Cotisations</h4>
              {!isAnonymized && <button className="btn btn--sm btn--ghost">+ Enregistrer un paiement</button>}
            </div>
            <table className="table">
              <thead>
                <tr><th>Saison</th><th>Montant</th><th>Date</th><th>Type</th><th>Commentaire</th><th></th></tr>
              </thead>
              <tbody>
                {MOCK.payments.map(p => (
                  <tr key={p.id}>
                    <td data-label="Saison"><strong>{p.season}</strong></td>
                    <td data-label="Montant" className="table__num">{p.amount.toFixed(2)} €</td>
                    <td data-label="Date">{p.date}</td>
                    <td data-label="Type">{p.type}</td>
                    <td data-label="Commentaire" className="muted">{p.comment || '—'}</td>
                    <td className="table__actions">
                      {!isAnonymized && <button className="btn btn--sm btn--ghost">Modifier</button>}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
            <div className="card__footer">
              <span className="muted">Total : {MOCK.payments.reduce((a,p)=>a+p.amount,0).toFixed(2)} € sur {MOCK.payments.length} saisons</span>
              <span className="muted">1ère inscription : 18 septembre 2018 (date du 1er paiement)</span>
            </div>
          </div>
        </div>

        {/* Admin sidebar */}
        <aside className="stack" style={{ '--gap': '16px' }}>
          {!isAnonymized && (
            <div className="card">
              <div className="card__header"><h4>Accès et sécurité</h4></div>
              <div className="card__body stack" style={{ '--gap': '12px' }}>
                {isPending ? (
                  <button className="btn btn--ghost" style={{ width: '100%', justifyContent: 'flex-start' }} onClick={() => setShowInvite(true)}>🔗 Générer un nouveau lien d'invitation</button>
                ) : (
                  <button className="btn btn--ghost" style={{ width: '100%', justifyContent: 'flex-start' }}>🔑 Générer un lien de réinitialisation</button>
                )}
                <button className="btn btn--ghost" style={{ width: '100%', justifyContent: 'flex-start' }} onClick={() => setShowBlocked(true)}>
                  {true ? '− Retirer le rôle admin' : '+ Accorder le rôle admin'}
                </button>
                <button className="btn btn--ghost" style={{ width: '100%', justifyContent: 'flex-start' }}>✗ Retirer le consentement téléphone/adresse</button>
                <button
                  className="btn btn--ghost"
                  style={{ width: '100%', justifyContent: 'flex-start' }}
                  onClick={() => setRestricted(r => !r)}
                >
                  {restricted ? '◎ Lever la restriction (art. 18 RGPD)' : '⊘ Restreindre le traitement (art. 18 RGPD)'}
                </button>
              </div>
            </div>
          )}
          <div className="card card--quiet">
            <div className="card__body" style={{ fontSize: 13, color: 'var(--ink-3)' }}>
              <strong style={{ color: 'var(--ink-2)', display: 'block', marginBottom: 6 }}>Activité</strong>
              <div>Dernière connexion : 19 avril 2026, 21:14</div>
              <div style={{ marginTop: 4 }}>Compte créé : 12 septembre 2018</div>
              <div style={{ marginTop: 4 }}>Rôle admin depuis : 3 janvier 2024</div>
            </div>
          </div>
        </aside>
      </div>

      {/* Modals */}
      {showInvite && <InviteLinkModal onClose={() => setShowInvite(false)} />}
      {showAnonymize && <AnonymizeModal onClose={() => setShowAnonymize(false)} />}
      {showBlocked && <LastAdminBlockModal onClose={() => setShowBlocked(false)} />}

      <SpecNote show={annotations}>
        <span>Les champs téléphone et adresse sont affichés uniquement si le consentement est donné.</span>
        {accountState === 'active' && <span>Anonymisation : irréversible, efface les champs identifiants, remplace les références dans les cotisations par un jeton opaque, supprime les RSVP. Bloquée si dernier admin.</span>}
        {isAnonymized && <span>L'instrument principal est conservé pour les statistiques. Le statut <code>anonymized</code> est terminal.</span>}
      </SpecNote>
    </div>
  );
}

function InviteLinkModal({ onClose }) {
  return (
    <div className="backdrop" onClick={onClose}>
      <div className="modal" onClick={e => e.stopPropagation()}>
        <div className="modal__header"><h3>Lien d'invitation</h3></div>
        <div className="modal__body stack" style={{ '--gap': '16px' }}>
          <p className="muted" style={{ margin: 0, fontSize: 14 }}>Copiez ce lien et envoyez-le manuellement au musicien (e-mail personnel, SMS). Valide 7 jours.</p>
          <div className="copy-link">
            <span className="copy-link__url">https://ohm-montrouge.fr/invitation/8a7f2e91c0b4d5a6f3e8c1b2a9d7f4e5</span>
            <button className="btn btn--sm btn--primary">Copier</button>
          </div>
          <div className="alert alert--info">
            <span className="alert__icon">ℹ</span>
            <div style={{ fontSize: 13 }}>Générer un nouveau lien invalidera le lien précédent. Le compte reste en <code>pending</code> tant que le musicien n'a pas complété son inscription.</div>
          </div>
        </div>
        <div className="modal__footer">
          <button className="btn btn--ghost" onClick={onClose}>Fermer</button>
          <button className="btn btn--primary">Régénérer</button>
        </div>
      </div>
    </div>
  );
}

function AnonymizeModal({ onClose }) {
  const [typed, setTyped] = useState('');
  return (
    <div className="backdrop" onClick={onClose}>
      <div className="modal" onClick={e => e.stopPropagation()}>
        <div className="modal__header" style={{ background: 'var(--danger-bg)', borderBottom: '1px solid #e8b7ac' }}>
          <h3 style={{ color: 'var(--danger)' }}>⚠ Anonymisation du compte</h3>
        </div>
        <div className="modal__body stack" style={{ '--gap': '14px' }}>
          <p style={{ margin: 0 }}>Cette action est <strong>irréversible</strong>. Elle satisfait une demande d'effacement RGPD (Art. 17).</p>
          <ul style={{ margin: 0, paddingLeft: 20, fontSize: 14, color: 'var(--ink-2)' }}>
            <li>Nom, e-mail, téléphone, adresse, mot de passe → <strong>effacés</strong></li>
            <li>Rôles supprimés · jetons invalidés · RSVP supprimés</li>
            <li>Cotisations : conservées (pseudonymisées avec un jeton opaque)</li>
            <li>Instrument principal : conservé (statistiques)</li>
          </ul>
          <div className="field">
            <label className="field__label">Pour confirmer, tapez « <code>ANONYMISER</code> »</label>
            <input className="input" value={typed} onChange={e => setTyped(e.target.value)} />
          </div>
        </div>
        <div className="modal__footer">
          <button className="btn btn--ghost" onClick={onClose}>Annuler</button>
          <button className="btn btn--danger" disabled={typed !== 'ANONYMISER'}>Anonymiser définitivement</button>
        </div>
      </div>
    </div>
  );
}

function LastAdminBlockModal({ onClose }) {
  return (
    <div className="backdrop" onClick={onClose}>
      <div className="modal" onClick={e => e.stopPropagation()}>
        <div className="modal__header" style={{ background: 'var(--warn-bg)', borderBottom: '1px solid #e6d09a' }}>
          <h3 style={{ color: 'var(--warn)' }}>🛡 Action bloquée — dernier administrateur</h3>
        </div>
        <div className="modal__body">
          <p style={{ marginTop: 0 }}>Cette opération laisserait zéro compte administrateur actif. Pour procéder, accordez d'abord le rôle admin à un autre compte actif.</p>
          <div className="alert alert--info">
            <span className="alert__icon">ℹ</span>
            <span style={{ fontSize: 13 }}>Invariant I3 — au moins un compte <code>active</code> doit porter le rôle admin à tout moment.</span>
          </div>
        </div>
        <div className="modal__footer"><button className="btn btn--primary" onClick={onClose}>Compris</button></div>
      </div>
    </div>
  );
}

// ============================================================
// NEW MUSICIAN FORM
// ============================================================
function AdminMusicianNewScreen({ navigate, annotations }) {
  const [hasConsent] = useState(false);
  const [under15] = useState(false);
  return (
    <div className="page page--narrow">
      <Breadcrumb items={[{ label: 'Musiciens', route: 'admin-musicians' }, { label: 'Nouveau' }]} navigate={navigate}/>
      <div className="page__header">
        <div className="page__title"><h1>Nouveau musicien</h1><p>Un lien d'invitation sera généré automatiquement à partager manuellement.</p></div>
      </div>

      <form className="card">
        <div className="card__body stack" style={{ '--gap': '20px' }}>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
            <div className="field">
              <label className="field__label">Prénom <span className="req">*</span></label>
              <input className="input" placeholder="Hugo"/>
            </div>
            <div className="field">
              <label className="field__label">Nom <span className="req">*</span></label>
              <input className="input" placeholder="Bernard"/>
            </div>
          </div>
          <div className="field">
            <label className="field__label">Adresse e-mail <span className="req">*</span></label>
            <input className="input" type="email" placeholder="hugo.bernard@example.com"/>
            <div className="field__hint">Doit être unique parmi les comptes non anonymisés. Sert d'identifiant de connexion.</div>
          </div>
          <div className="field">
            <label className="field__label">Instrument principal <span className="req">*</span></label>
            <select className="select">{MOCK.instruments.map(i => <option key={i}>{i}</option>)}</select>
          </div>
          <div className="field">
            <label className="field__label">Date de naissance</label>
            <input className="input" type="date"/>
            <div className="field__hint">Optionnel. Si le musicien a moins de 15 ans, un URI de consentement parental devient obligatoire.</div>
          </div>
          {under15 && (
            <div className="field">
              <label className="field__label">URI de consentement parental <span className="req">*</span></label>
              <input className="input" placeholder="https://drive.google.com/..."/>
              <div className="field__hint">Document attestant du consentement parental (Art. 8 RGPD).</div>
            </div>
          )}
          <div>
            <div className="field__label" style={{ marginBottom: 8 }}>Téléphone et adresse</div>
            <div className="alert alert--warn" style={{ padding: '10px 14px' }}>
              <span className="alert__icon">🔒</span>
              <div style={{ fontSize: 13 }}>Ces champs seront verrouillés jusqu'à ce que le musicien donne son consentement à la fin du flux d'invitation.</div>
            </div>
          </div>
        </div>
        <div className="card__footer">
          <button className="btn btn--ghost" onClick={() => navigate('admin-musicians')}>Annuler</button>
          <button className="btn btn--primary">Créer le compte et générer l'invitation</button>
        </div>
      </form>

      <SpecNote show={annotations}>
        <span>Téléphone et adresse ne sont modifiables par l'admin qu'après consentement du musicien (donné dans le flux d'invitation).</span>
        <span>Règle des moins de 15 ans : appliquée à chaque sauvegarde.</span>
      </SpecNote>
    </div>
  );
}

Object.assign(window, { AdminMusiciansScreen, AdminMusicianDetailScreen, AdminMusicianNewScreen });
