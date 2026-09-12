export const routes = {
  home: "/",
  login: "/login",
  register: "/register",
  notification: "/notification",
  settings: "/userprofile",

  artist: {
    profile: "/artist-profile",
  },

  order: {
    history: "/order-history",
    detail: (id: string) => `/order-history/${id}`,
  },
} as const;