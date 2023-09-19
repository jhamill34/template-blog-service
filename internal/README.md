The heart of our application, i.e., all its internal logic, is stored here. /internal is not imported into other applications and libraries. The code written here is intended solely for internal use within the code base. Starting from the Go 1.4 version, a defined mechanism has been in place that prevents the importing of packages outside of this project if they are inside /internal.

/internal is where we store the business logic of the project, together with all work related to databases. In other words, all the logic associated with this app. The structure inside /internal can be organized in a variety of ways, depending on the specific architecture used. I’m not going to go into this in too much depth right now, but I will illustrate in broad strokes what it looks like. Here is an example of a three-layer architecture, where the app is divided into three layers:

Transport.
Business.
Databases.

The logic should be such that the layers hierarchically communicate with each other from top to bottom, and vice versa. No layer may “skip” over its intermediate peer (e.g. when the transport layer communicates directly with the database) and no layer that is below another one may communicate directly with the layer above (e.g. when the database communicates with the transport layer).
