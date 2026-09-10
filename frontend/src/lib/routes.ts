export const routes = {
  home: "/",
  login: "/login",
  register: "/register",
  notification: "/notification",
  settings: "/userprofile",

  artist: {
    profile: "/artistprofile",
  },

  order: {
    history: "/orderhistory",
    detail: (id: string) => `/orderhistory/${id}`,
  },
} as const;
