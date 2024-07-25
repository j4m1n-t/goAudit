# TODO
<ul>
<li>Encrypt passwords</li>
<li>Setup the master password functionality</li>
<li>Update:</li>
   <ol> <li>Tasks</li>
   <li>CRM</li>
   <li>Audits</li>
</ol></ul>
<sub>They need to have a better looking interface.</sub>



## Decide on server / serverless 


|Server Pros|Server Cons|Serverless Pros|ServerLess Cons|
|:--:|:--:|:--:|:--:|
|Central setup for admins| More backend coding| Less backend coding | More work setting up for admins|
|Clients will be able to retrieve config from the server configuration| Expense | Savings | To share the configuration you would need to create a shared drive or similar|


## How we want this to work
Basic fuctionality
<ol>
<li>Login using ldap / active directory.</li>
<li>Backend creates a user and updates the db with username, user_id, and id.</li>
<li>System will check if admin to allow access to the admin tab.</li>
<li>System checks if user has a master password in "Credentials".</li>
<li>If yes, request master password for access to credentials tab.</li>
<li>If no, signup dialog for username (prefilled with username), master password, verify master password, and email.</li>
<li>Once fully authenticated users will be able to use all modules of the program.</li></ol>

Advanced Functionality
<ul>
<li>Mangagers are able to assign tasks to users based on ldap/ad permissions.*</li>
<li>Admins are able to create, edit, view, or delete anything EXCEPT credentials.</li>
</ul>

<sup>*This may be something we modify in the future to allow admins to grant access to managers with using ldap/ad</sup>