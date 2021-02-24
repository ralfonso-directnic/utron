package session


    import(
            "encoding/gob"
            "github.com/gorilla/sessions"
            "database/sql"
            "github.com/ralfonso-directnic/utron/config"
		    "github.com/gernest/qlstore"
	        _ "github.com/cznic/ql/driver"
	      // "github.com/gorilla/securecookie"
    )



    func init(){
        
        gob.Register(map[string]interface{}{})
        gob.Register(map[string]string{})
        
        
    }


    type Session struct {
    	config  *config.Config
    	Store sessions.Store
    }
    
    
    func New(cfg *config.Config) (*Session){
        
        return &Session{config:cfg}
        
    }
    
    
    
    
    //Loads the store set in the config
        
    func (s *Session) LoadStore() (sessions.Store,error) {


    
       var err error
       
    	
    	switch s.config.SessionStore {
        	
        	 case "file":
        	  s.Store,err = s.fileStore()
        	 case "cookie":
        	  s.Store,err = s.cookieStore()
        	 case "sqlite":
        	  s.Store,err = s.sqliteStore()
        	 default:
              s.Store,err = s.cookieStore()
        	     		
    	}
    	
    	
    	
     	return s.Store,err

	
	}
	
	func (s *Session) getOptions() (*sessions.Options){
    	
    	
    	opts := &sessions.Options{
    		Path:     s.config.SessionPath,
    		Domain:   s.config.SessionDomain,
    		MaxAge:   s.config.SessionMaxAge,
    		Secure:   s.config.SessionSecure,
    		HttpOnly: s.config.SessionSecure,
    	}
    	
    	return opts
        	
	}
	
	
	func (s *Session) sqliteStore() (sessions.Store,error){
    	
    
    	
    	opts := s.getOptions()
    	
        db, err := sql.Open("ql-mem", "session.db")
    	if err != nil {
    		return nil, err
    	}
    	err = qlstore.Migrate(db)
    	if err != nil {
    		return nil, err
    	}
    
    	store := qlstore.NewQLStore(db, "/", 2592000, keyPairs(s.config.SessionKeyPair)...)
    	store.Options = opts
    	
    	s.Store = store
    	return s.Store, nil
        	
	}
	
	func (s *Session) cookieStore() (sessions.Store,error){ 
    	

    	
    	opts := s.getOptions()
    	    	
        //authKeyOne := securecookie.GenerateRandomKey(64)
       // encryptionKeyOne := securecookie.GenerateRandomKey(32)



        store := sessions.NewCookieStore(
                keyPairs(s.config.SessionKeyPair)...
        )
        
        store.Options = opts

        s.Store = store

    	return s.Store,nil
    	
	}
	
    func (s *Session) fileStore() (sessions.Store,error){ 
    	
    	
        opts := s.getOptions()
    	
        //authKeyOne := securecookie.GenerateRandomKey(64)
        //encryptionKeyOne := securecookie.GenerateRandomKey(32)


    	store := sessions.NewFilesystemStore(
    		s.config.SessionFilePath,
    		keyPairs(s.config.SessionKeyPair)...
    	)

        store.Options = opts
    	
    	s.Store = store
    	
    	return s.Store,nil
    	
	}
	
	
	func keyPairs(src []string) [][]byte {
    	var pairs [][]byte
    	for _, v := range src {
    		pairs = append(pairs, []byte(v))
    	}
    	return pairs
    }
	
	