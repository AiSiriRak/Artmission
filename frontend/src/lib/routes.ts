export const routes = {
  home: "/",
  login: "/login",
  register: "/register",
  notification: "/notification",
<<<<<<< HEAD
  settings: "/userprofle",
=======
  settings: "/userprofile",
>>>>>>> feat/frontend/user-profile-page

  artist: {
    profile: "/artist-profile",
  },

  order: {
    history: "/order-history",
    detail: (id: string) => `/order-history/${id}`,
  },
} as const;
