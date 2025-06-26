import axios from 'axios'

export const useAxios = () => {
  return axios.create({
    timeout: 100000,
  })
}
