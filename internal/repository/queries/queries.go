package queries

////////////////////////////////////////
// usersrepo

const SelectUser = `
		SELECT users.userID
		FROM
			public.users
		WHERE
		users.userLogin=$1
	`

const SelectUserWithPass = `
		SELECT users.userID
		FROM
			public.users
		WHERE
		users.userLogin=$1 AND users.userPassword = $2
	`

const CreateUserInsert = `
		INSERT INTO public.users
		(userLogin, userPassword)
		VALUES
		($1, $2);
	`

////////////////////////////////////////
// secretsrepo

const SelectSecret = `
		SELECT secrets.id
		FROM
			public.secrets
		WHERE
		secrets.name=$1 AND secrets.ownerID = $2
	`

const CreateSecretInsert = `
	INSERT INTO public.secrets
	(name, type, content, ownerID, keyID)
	VALUES
	($1, $2, $3, $4, $5);
`
const SelectSecrets = `
		SELECT secrets.name, secrets.type
		FROM
			public.secrets
		WHERE
		secrets.ownerID = $1
	`
const SelectSecretInfo = `
	SELECT secrets.type, secrets.content, secrets.version, secrets.keyID
	FROM
		public.secrets
	WHERE
	secrets.name=$1 AND secrets.ownerID = $2
`
const DeleteSecret = `
	DELETE FROM public.secrets
	WHERE secrets.name=$1 AND secrets.ownerID = $2;
`

////////////////////////////////////////
// sync

const SyncDataKeys = `
	SELECT keyID, keyAES, ownerID, timestamp::text 
	FROM 
		public.keys 
	WHERE 
	ownerID = $1
`
const PushDataKeys = `
	INSERT INTO Keys (keyID, keyAES, ownerID, timestamp) 
	VALUES ($1, $2, $3, $4) 
	ON CONFLICT (keyID) 
	DO UPDATE SET KeyAES = EXCLUDED.KeyAES, timestamp = EXCLUDED.timestamp 
	WHERE EXCLUDED.timestamp > Keys.timestamp
`
const SyncDataUsers = `
	SELECT userID, userLogin, userPassword 
	FROM 
		public.users 
	WHERE 
	userID = $1
`
const SyncDataSecrets = `
	SELECT name, type, content, ownerID, keyID, timestamp::text, isDeleted  
	FROM 
		public.secrets 
	WHERE 
	ownerID = $1
`
const PushDataSecrets = `
	INSERT INTO Secrets (name, type, content, ownerID, keyID, timestamp, isDeleted) 
	VALUES ($1, $2, $3, $4, $5, $6, $7) 
	WHERE EXCLUDED.timestamp > Secrets.timestamp
`
