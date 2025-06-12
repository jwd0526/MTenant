CREATE TABLE deal_contacts (
   deal_id INTEGER REFERENCES deals(id) ON DELETE CASCADE,
   contact_id INTEGER REFERENCES contacts(id) ON DELETE CASCADE,
   role VARCHAR(100),
   created_at TIMESTAMP DEFAULT NOW(),
   PRIMARY KEY(deal_id, contact_id)
);

CREATE INDEX idx_deal_contacts_deal ON deal_contacts(deal_id);
CREATE INDEX idx_deal_contacts_contact ON deal_contacts(contact_id);