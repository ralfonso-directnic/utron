package base

import (
	"errors"

	"github.com/gorilla/sessions"
)

var errNoStore = errors.New("no session store was found")


//GetSession retrieves session with a given name.
func (ctx *Context) GetSession() (*sessions.Session, error) {
	
	name := ctx.Cfg.SessionName
	if ctx.SessionStore != nil {
		return ctx.SessionStore.New(ctx.Request(), name)
	}
	return nil, errNoStore
}

//SaveSession saves the given session.
func (ctx *Context) SaveSession(s *sessions.Session) error {
	if ctx.SessionStore != nil {
		return ctx.SessionStore.Save(ctx.Request(), ctx.Response(), s)
	}
	return errNoStore
}


func (ctx *Context) GetSessionValues() (map[interface{}]interface{}) {
    
       sess,_ := ctx.GetSession()
    
       return sess.Values
    
}


func (ctx *Context) SessionSet(key interface{},val interface{},save ...bool) (error){
    
    
    sess,err := ctx.GetSession()
    
    
    if(err!=nil){
        
        return err
    }
    
    sess.Values[key]=val
    
    
    for _, sv := range save {        
        
        if(sv==true){
            
            ctx.SaveSession(sess)
        }
        
        break
        
    }
    
    return nil
    
    
    
}

func (ctx *Context) SessionGet(key interface{}) (interface{},error) {
    
    sess,err := ctx.GetSession()
    
    if(err!=nil){
        
        return nil,err
    }
    
    
    
    if val,ok := sess.Values[key]; ok {
        
        return val,nil
        
    }
    
    return nil,err
    
    
}




