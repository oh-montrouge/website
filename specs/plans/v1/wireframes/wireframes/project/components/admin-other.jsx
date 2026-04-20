/* global React, MOCK, fmtDate, fmtTime, EventTypeBadge, SpecNote, Breadcrumb */
const { useState } = React;

// ============================================================
// ADMIN EVENTS LIST
// ============================================================
function AdminEventsScreen({ navigate, annotations }) {
  return (
    <div className="page page--wide">
      <div className="page__header">
        <div className="page__title"><h1>Gestion des événements</h1><p>Concerts, répétitions et autres rendez-vous.</p></div>
        <div className="page__actions">
          <button className="btn btn--primary" onClick={() => navigate('admin-event-edit')}>+ Nouvel événement</button>
        </div>
      </div>
      <div className="card" style={{ overflow: 'hidden' }}>
        <table className="table">
          <thead><tr><th>Date</th><th>Nom</th><th>Type</th><th>RSVP</th><th></th></tr></thead>
          <tbody>
            {MOCK.events.map(ev => (
              <tr key={ev.id}>
                <td data-label="Date"><strong>{fmtDate(ev.date)}</strong><div className="muted" style={{ fontSize: 12 }}>{fmtTime(ev.date)}</div></td>
                <td data-label="Nom"><strong>{ev.name}</strong></td>
                <td data-label="Type"><EventTypeBadge type={ev.type}/></td>
                <td data-label="Réponses" className="muted">5 oui · 1 peut-être · 1 non · 2 sans réponse</td>
                <td className="table__actions">
                  <button className="btn btn--sm btn--ghost" onClick={() => navigate('event-detail', { eventType: ev.type })}>Voir</button>
                  <button className="btn btn--sm btn--ghost" onClick={() => navigate('admin-event-edit', { eventType: ev.type })}>Modifier</button>
                  <button className="btn btn--sm btn--danger">Supprimer</button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      <SpecNote show={annotations}>
        <span>Supprimer un événement efface tous ses RSVP — c'est le chemin RGPD pour la purge des données de présence.</span>
        <span>À la création : un RSVP (état <code>unanswered</code>) est créé pour chaque compte <code>active</code>.</span>
      </SpecNote>
    </div>
  );
}

// ============================================================
// EVENT EDIT FORM + custom fields panel
// ============================================================
function AdminEventEditScreen({ navigate, annotations, eventType: initType = 'other' }) {
  const [type, setType] = useState(initType);
  const [fields, setFields] = useState([
    { id: 1, label: 'Covoiturage', type: 'choice', required: true,  choices: ['Je propose', 'Je cherche', 'Pas besoin'] },
    { id: 2, label: 'Repas',       type: 'choice', required: true,  choices: ['Standard', 'Végétarien', 'Végan'] },
    { id: 3, label: "Heure d'arrivée", type: 'text', required: false, choices: [] },
  ]);

  return (
    <div className="page">
      <Breadcrumb items={[{ label: 'Événements (admin)', route: 'admin-events' }, { label: 'Week-end de travail' }]} navigate={navigate}/>
      <div className="page__header">
        <div className="page__title"><h1>Modifier l'événement</h1></div>
        <div className="page__actions">
          <button className="btn btn--ghost" onClick={() => navigate('admin-events')}>Annuler</button>
          <button className="btn btn--primary">Enregistrer</button>
        </div>
      </div>

      <div className="split">
        <form className="card">
          <div className="card__body stack" style={{ '--gap': '20px' }}>
            <div className="field">
              <label className="field__label">Nom <span className="req">*</span></label>
              <input className="input" defaultValue="Week-end de travail"/>
            </div>
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
              <div className="field">
                <label className="field__label">Date <span className="req">*</span></label>
                <input className="input" type="date" defaultValue="2026-06-07"/>
              </div>
              <div className="field">
                <label className="field__label">Heure <span className="req">*</span></label>
                <input className="input" type="time" defaultValue="09:00"/>
              </div>
            </div>
            <div className="field">
              <label className="field__label">Type <span className="req">*</span></label>
              <div className="radio-group">
                {['concert','rehearsal','other'].map(t => (
                  <button type="button" key={t} className={`radio-pill ${type===t?'radio-pill--active':''}`} onClick={() => setType(t)}>
                    {t==='concert'?'Concert':t==='rehearsal'?'Répétition':'Autre'}
                  </button>
                ))}
              </div>
              {type !== 'other' && (
                <div className="alert alert--warn" style={{ marginTop: 12 }}>
                  <span className="alert__icon">⚠</span>
                  <div style={{ fontSize: 13 }}>
                    Passer à <strong>{type === 'concert' ? 'Concert' : 'Répétition'}</strong> depuis « Autre » supprimera tous les champs personnalisés et réinitialisera les RSVP « oui ».
                  </div>
                </div>
              )}
            </div>
          </div>
        </form>

        <aside>
          {type === 'other' && (
            <div className="card">
              <div className="card__header">
                <h4>Champs personnalisés</h4>
                <button className="btn btn--sm">+ Ajouter</button>
              </div>
              <div className="card__body stack" style={{ '--gap': '12px' }}>
                {fields.map(f => (
                  <div key={f.id} style={{ border: '1px solid var(--line)', borderRadius: 6, padding: '10px 12px' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                      <div>
                        <strong style={{ fontSize: 14 }}>{f.label}</strong>
                        <div className="muted" style={{ fontSize: 12 }}>
                          {f.type} {f.required && '· obligatoire'} {f.choices.length > 0 && `· ${f.choices.length} choix`}
                        </div>
                      </div>
                      <div className="row" style={{ gap: 4 }}>
                        <button className="btn btn--sm btn--ghost">✎</button>
                        <button className="btn btn--sm btn--ghost" style={{ color: 'var(--danger)' }}>×</button>
                      </div>
                    </div>
                  </div>
                ))}
                <div className="alert alert--info" style={{ padding: '10px 12px', fontSize: 12 }}>
                  <span className="alert__icon">ℹ</span>
                  <span>Les champs ne peuvent être modifiés ou supprimés qu'avant la première réponse enregistrée.</span>
                </div>
              </div>
            </div>
          )}
          {type !== 'other' && (
            <div className="card card--quiet">
              <div className="card__body muted" style={{ fontSize: 13 }}>
                Les champs personnalisés ne sont disponibles que pour les événements de type « Autre ».
              </div>
            </div>
          )}
        </aside>
      </div>

      <SpecNote show={annotations}>
        <span>Changement de type : effets précis sur RSVP/champs (voir table 05-events-and-rsvp §Editing an Event).</span>
        <span>Autre → Concert/Répétition : champs et réponses supprimés.</span>
      </SpecNote>
    </div>
  );
}

// ============================================================
// SEASONS
// ============================================================
function AdminSeasonsScreen({ navigate, annotations }) {
  const [showNew, setShowNew] = useState(false);
  return (
    <div className="page">
      <div className="page__header">
        <div className="page__title"><h1>Saisons</h1><p>Une saison correspond à une période de cotisation annuelle. Immuable après création.</p></div>
        <div className="page__actions">
          <button className="btn btn--primary" onClick={() => setShowNew(true)}>+ Nouvelle saison</button>
        </div>
      </div>
      <div className="card" style={{ overflow: 'hidden' }}>
        <table className="table">
          <thead><tr><th>Libellé</th><th>Début</th><th>Fin</th><th>Cotisations</th><th></th></tr></thead>
          <tbody>
            {MOCK.seasons.map(s => (
              <tr key={s.id}>
                <td data-label="Saison">
                  <strong>{s.label}</strong>
                  {s.current && <span className="badge badge--ok" style={{ marginLeft: 10 }}>● Saison courante</span>}
                </td>
                <td data-label="Début">{s.start}</td>
                <td data-label="Fin">{s.end}</td>
                <td data-label="Paiements" className="muted">{s.payments} paiements</td>
                <td className="table__actions">
                  {!s.current && <button className="btn btn--sm btn--ghost">Désigner comme courante</button>}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      {showNew && (
        <div className="backdrop" onClick={() => setShowNew(false)}>
          <div className="modal" onClick={e => e.stopPropagation()}>
            <div className="modal__header"><h3>Nouvelle saison</h3></div>
            <div className="modal__body stack" style={{ '--gap': '14px' }}>
              <div className="field"><label className="field__label">Libellé <span className="req">*</span></label><input className="input" placeholder="2026-2027"/></div>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
                <div className="field"><label className="field__label">Début <span className="req">*</span></label><input className="input" type="date"/></div>
                <div className="field"><label className="field__label">Fin <span className="req">*</span></label><input className="input" type="date"/></div>
              </div>
              <div className="alert alert--info"><span className="alert__icon">ℹ</span><span style={{ fontSize: 13 }}>La nouvelle saison ne sera pas désignée courante automatiquement.</span></div>
            </div>
            <div className="modal__footer">
              <button className="btn btn--ghost" onClick={() => setShowNew(false)}>Annuler</button>
              <button className="btn btn--primary">Créer</button>
            </div>
          </div>
        </div>
      )}
      <SpecNote show={annotations}>
        <span>Invariant I4 : exactement une saison est courante à tout moment. Changer la courante transfère la désignation ; l'ancienne reste et accepte encore des paiements.</span>
        <span>Les saisons sont immuables après création.</span>
      </SpecNote>
    </div>
  );
}

// ============================================================
// RETENTION
// ============================================================
function AdminRetentionScreen({ navigate, annotations }) {
  const items = [
    { name: 'Pierre Lambert', instrument: 'Trombone', lastSeason: '2019-2020', endDate: '31 août 2020' },
    { name: 'Nathalie Roux',  instrument: 'Flûte',    lastSeason: '2018-2019', endDate: '31 août 2019' },
    { name: 'Michel Dupont',  instrument: 'Tuba',     lastSeason: '2020-2021', endDate: '31 août 2021' },
  ];
  return (
    <div className="page">
      <div className="page__header">
        <div className="page__title">
          <h1>Revue de rétention</h1>
          <p>Comptes dont la dernière cotisation est terminée depuis plus de 5 ans. À examiner pour anonymisation.</p>
        </div>
      </div>
      <div className="alert alert--info" style={{ marginBottom: 24 }}>
        <span className="alert__icon">ℹ</span>
        <div style={{ fontSize: 14 }}>
          Le système ne supprime rien automatiquement. Vérifiez chaque dossier avant d'anonymiser. Les comptes sans cotisation n'apparaissent pas ici.
        </div>
      </div>
      <div className="card" style={{ overflow: 'hidden' }}>
        <table className="table">
          <thead><tr><th>Musicien</th><th>Instrument</th><th>Dernière saison cotisée</th><th>Fin de saison</th><th></th></tr></thead>
          <tbody>
            {items.map((i, idx) => (
              <tr key={idx}>
                <td data-label="Musicien"><strong>{i.name}</strong></td>
                <td data-label="Instrument">{i.instrument}</td>
                <td data-label="Dernière saison"><span className="badge badge--warn">{i.lastSeason}</span></td>
                <td data-label="Fin de saison" className="muted">{i.endDate}</td>
                <td className="table__actions">
                  <button className="btn btn--sm btn--ghost">Voir la fiche</button>
                  <button className="btn btn--sm btn--danger">Anonymiser</button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      <SpecNote show={annotations}>
        <span>Critères : ≥ 1 cotisation ET fin de dernière saison cotisée &gt; 5 ans ET non anonymisé.</span>
      </SpecNote>
    </div>
  );
}

Object.assign(window, { AdminEventsScreen, AdminEventEditScreen, AdminSeasonsScreen, AdminRetentionScreen });
