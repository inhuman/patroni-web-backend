#Partoni web backend

Requirements:

 - Postrge Db
 - Consul
 - FreeIPA
 - ELK (not required) 
 
 
Authorization:

 For auth user must have account in FreeIpa
 
 To be an admin on web ui, user must have group "patroni-admins"
 
 For edit clusters user must have group "patroni-dc-<CLUSTER_NAME>"
 